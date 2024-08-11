package client

import (
	"context"
	"errors"
	"github.com/injoyai/base/bytes"
	"github.com/injoyai/base/maps"
	"github.com/injoyai/base/safe"
	"github.com/injoyai/ios"
	"time"
)

type (
	Option  func(c *Client)
	Message = bytes.Entity
)

func Run(f ios.DialFunc, op ...Option) error {
	c, err := Dial(f, op...)
	if err != nil {
		return err
	}
	return c.Run()
}

func MustDial(f ios.DialFunc, op ...Option) *Client {
	return MustDialWithContext(context.Background(), f, op...)
}

func MustDialWithContext(ctx context.Context, f ios.DialFunc, op ...Option) *Client {
	c := newClient(ctx, f, op...)
	_ = c.connect(true)
	return c
}

func Dial(f ios.DialFunc, op ...Option) (*Client, error) {
	return DialWithContext(context.Background(), f, op...)
}

func DialWithContext(ctx context.Context, f ios.DialFunc, op ...Option) (*Client, error) {
	c := newClient(ctx, f, op...)
	err := c.connect(false)
	return c, err
}

func WithMust(h ios.DialFunc, l Logger) ios.DialFunc {
	return func(ctx context.Context) (ios.ReadWriteCloser, string, error) {
		if h == nil {
			return nil, "", errors.New("handler is nil")
		}
		wait := time.Second * 0
		for i := 0; ; i++ {
			select {
			case <-ctx.Done():
				return nil, "", ctx.Err()
			case <-time.After(wait):
				c, k, err := h(ctx)
				if err == nil {
					return c, k, nil
				}
				if wait < time.Second {
					wait = time.Second
				} else if wait <= time.Second*16 {
					wait *= 2
				}
				if l != nil {
					l.Errorf("连接错误: %v 等待%d秒后重试...\n", err, wait/time.Second)
				}
			}
		}
	}
}

func newClient(ctx context.Context, f ios.DialFunc, op ...Option) *Client {
	return &Client{
		Tag:    maps.NewSafe(),
		ctx:    ctx,
		Logger: defaultLogger,
		Info: Info{
			CreateTime: time.Now(),
			DialTime:   time.Now(),
		},
		Event:   &Event{},
		dial:    f,
		options: op,
	}
}

/*
Client
客户端的指针地址是唯一标识,key是表面的唯一标识,需要用户自己维护
*/
type Client struct {
	Key            string //自定义标识
	ios.Reader            //IO实例
	ios.MoreWriter        //多个方式写入

	Tag          *maps.Safe //标签,用于记录连接的一些信息
	Logger                  //日志
	Info                    //基本信息
	*Event                  //事件
	*safe.Closer            //关闭
	*safe.Runner            //运行

	ctx         context.Context //上下文
	readBuffer  []byte          //读数据的缓存大小,针对io.Reader有效
	redial      bool            //是否重连
	dial        ios.DialFunc    //连接函数
	options     []Option        //选项
	readTimeout time.Duration   //读取超时时间
	timeout     *safe.Runner
}

func (this *Client) connect(must bool) (err error) {

	defer func() { this.CloseWithErr(err) }()

	f := this.dial
	if must {
		f = WithMust(f, this.Logger)
	}

	r, k, err := f(this.ctx)
	if err != nil {

		return err
	}

	this.Key = k
	this.Reader = r
	this.MoreWriter = ios.NewMoreWriter(r)
	this.Info.DialTime = time.Now()
	this.Runner = safe.NewRunnerWithContext(this.ctx, this.run)
	this.timeout = safe.NewRunnerWithContext(this.ctx, func(ctx context.Context) error {
		if this.readTimeout <= 0 {
			return nil
		}
		timer := time.NewTimer(this.readTimeout)
		defer timer.Stop()
		for {
			timer.Reset(this.readTimeout)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timer.C:
				return ios.ErrReadTimeout
			}
		}
	})
	this.Closer = safe.NewCloser().SetCloseFunc(func(err error) error {
		//关闭真实实例
		if er := r.Close(); er != nil {
			return er
		}
		this.Logger.Errorf("[%s] 断开连接: %s\n", this.GetKey(), err.Error())
		//结束Run,不等待,重连是直接在原先的Run里面
		this.Runner.Stop()
		this.timeout.Stop()
		//关闭/断开连接事件
		if this.Event.OnDisconnect != nil {
			this.Event.OnDisconnect(this, err)
		}
		return nil
	})
	//this.Logger
	//this.Event
	//this.tag

	//连接事件
	this.Logger.Infof("[%s] 连接服务成功...\n", this.GetKey())
	if this.Event.OnConnect != nil {
		if err := this.Event.OnConnect(this); err != nil {
			return err
		}
	}

	//写入事件
	this.MoreWriter.(*ios.MoreWrite).Option = []ios.WriteOption{
		func(p []byte) ([]byte, error) {
			this.Logger.Writeln("["+this.GetKey()+"] ", p)
			if this.Event.OnWriteMessage != nil {
				return this.Event.OnWriteMessage(p)
			}
			return p, nil
		},
	}

	//执行选项
	this.SetOption(this.options...)

	return nil
}

func (this *Client) SetReadTimeout(t time.Duration) *Client {
	this.readTimeout = t
	this.timeout.Restart()
	return this
}

func (this *Client) SetOption(op ...Option) *Client {
	for _, fn := range op {
		fn(this)
	}
	return this
}

func (this *Client) Redial(b ...bool) *Client {
	this.redial = len(b) == 0 || b[0]
	return this
}

// SetReadBuffer 设置读取缓存的大小,只针对io.Reader有效
func (this *Client) SetReadBuffer(size int) *Client {
	this.readBuffer = make([]byte, size)
	return this
}

func (this *Client) GetKey() string {
	return this.Key
}

func (this *Client) SetKey(key string) *Client {
	oldKey := this.Key
	this.Key = key
	if this.Key != oldKey && this.Event.OnKeyChange != nil {
		this.Event.OnKeyChange(this, oldKey)
	}
	return this
}

func (this *Client) Timer(t time.Duration, f Option) {
	tick := time.NewTicker(t)
	defer tick.Stop()
	for {
		select {
		case <-this.Closer.Done():
			return
		case _, ok := <-tick.C:
			if ok {
				f(this)
			}
		}
	}
}

func (this *Client) TimerWriter(t time.Duration, f func(w ios.MoreWriter) error) {
	this.Timer(t, func(c *Client) {
		c.CloseWithErr(f(c))
	})
}

func (this *Client) CloseAll() error {
	this.redial = false
	return this.Closer.Close()
}

func (this *Client) run(ctx context.Context) (err error) {

	this.timeout.Start()

	for {
		select {
		case <-ctx.Done():
			//这个ctx不是Client的ctx,而是Runner的ctx属于Client的ctx的子级
			if this.redial {
				//设置了重连,并且已经运行,其他都关闭
				if err := this.connect(true); err != nil {
					return err
				}
				return this.Run()
			}
			return this.Closer.Err()

		default:

			//校验事件函数
			if this.Event == nil {
				this.Event = &Event{}
			}
			if this.Event.OnReadBuffer == nil {
				this.Event.OnReadBuffer = ios.ReadBuffer
			}
			if this.Event.OnDealMessage != nil {
				this.Event.OnDealMessage = func(c *Client, message ios.Acker) {}
			}
			if this.readBuffer == nil {
				this.readBuffer = make([]byte, 1024)
			}

			//读取数据
			ack, err := this.Event.OnReadBuffer(this.Reader, this.readBuffer)
			if err != nil {
				this.CloseWithErr(err)
				continue
				return err
			}
			this.Info.ReadTime = time.Now()
			this.Logger.Readln("["+this.GetKey()+"] ", ack.Payload())

			//处理数据
			this.Event.OnDealMessage(this, ack)

		}
	}

}
