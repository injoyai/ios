package client

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/injoyai/base/maps"
	"github.com/injoyai/base/safe"
	"github.com/injoyai/ios"
	"github.com/injoyai/ios/module/common"
)

func Run(f ios.DialFunc, op ...Option) error {
	c, err := Dial(f, op...)
	if err != nil {
		return err
	}
	return c.Run(context.Background())
}

func Redial(f ios.DialFunc, op ...Option) *Client {
	return RedialContext(context.Background(), f, op...)
}

func RedialContext(ctx context.Context, dial ios.DialFunc, op ...Option) *Client {
	c := New(dial, op...)
	c.SetRedial()
	_ = c.Dial(ctx)
	return c
}

func Dial(f ios.DialFunc, op ...Option) (*Client, error) {
	return DialContext(context.Background(), f, op...)
}

func DialContext(ctx context.Context, dial ios.DialFunc, op ...Option) (*Client, error) {
	c := New(dial, op...)
	err := c.Dial(ctx)
	return c, err
}

func New(dial ios.DialFunc, op ...Option) *Client {
	c := &Client{}
	c.Reset()
	c.SetDial(dial)
	//这里直接执行Option,
	//如果想Write,则使用OnConnect进行处理,
	//否则设置重试等Option会无效
	c.SetOption(op...)
	return c
}

/*
Client
客户端的指针地址是唯一标识,key是表面的唯一标识,需要用户自己维护
*/
type Client struct {

	//IO实例,原始数据
	r ios.ReadWriteCloser

	//带缓存的reader,目前支持ios.AReader,ios.MReader,io.Reader
	buf ios.Reader

	//全局自定义标识,表明客户端的身份
	//默认使用的是客户端的IP:PORT
	//例如,通过解析注册信息后,可以使用解析的IMEI等信息作为key
	key string

	//是否自动重连
	//当连接断开的时候,自动重连,使用递归的方式
	//还是同一个客户端
	redial bool

	//实现多种读取方式
	//包括 io.Reader,ios.AReader,ios.MReader
	ios.AllReader

	//多个方式写入的封装
	//包括 Writer,StringWriter,ByteWriter等
	ios.MoreWriter

	//基本信息,一些连接时间,数据时间,数据大小等数据
	Info

	//各种事件,连接成功事件,数据读取(分包)事件,数据处理事件,连接关闭事件等
	//由用户自行配置,如果必须的事件未设置,则使用默认值
	//例如未设置读取(分包)事件,则默认使用一次读取最多4KB,能满足绝大部分需求
	*Event

	//安全关闭,单次的生命周期,每次重连都会重新声明
	*safe.Closer

	//运行,全局的生命周期,包括重试
	*safe.Runner2

	//日志管理,默认使用common.NewLogger
	Logger common.Logger

	//标签,用于自定义记录连接的一些信息
	//例如,客户端的ICCID,IMEI等
	Tag *maps.Safe

	//超时机制,监听客户端的读取和写入数据,维持不超时
	timeout *safe.Runner2

	//全局重连信号,未设置自动重连也可以手动重连,
	//向这个通道发送一个信号,则客户端会进行断开重连
	redialSign chan struct{}

	//全局已连接信号,仅在连接成功的时候会向这个通道发送20个信号,
	//便于一些逻辑的判断,如果改成单个生命周期的监听,会改变底层指针,不适用
	dialedSign chan struct{}

	//缓存连接函数,重连的时候使用
	dial ios.DialFunc

	//缓存选项,重连的时候使用
	options []Option
}

// Reset 重置参数,方便配合sync.Pool使用
func (this *Client) Reset() {
	this.key = ""
	this.r = nil
	this.buf = nil
	this.AllReader = nil
	this.MoreWriter = nil
	this.Logger = common.NewLogger()
	this.Info = Info{
		CreateTime: time.Now(),
	}
	this.Event = &Event{
		OnDealErr: func(c *Client, err error) error { return common.DealErr(err) },
	}
	this.Closer = safe.NewCloser()
	this.Runner2 = safe.NewRunner2(nil)
	this.Tag = maps.NewSafe()

	this.timeout = safe.NewRunner2(nil)
	this.redial = false
	this.redialSign = make(chan struct{})
	this.dialedSign = make(chan struct{})
	this.dial = nil
	this.options = nil

	this.Closer.CloseWithErr(errors.New("等待连接"))
	this.Runner2.SetFunc(this.run)
}

// Origin 获取原始连接,
// 尽量不要直接进行读取,
// 因为内部封装了一层buffer,可能会造成数据混乱
func (this *Client) Origin() ios.ReadWriteCloser {
	return this.r
}

// initAllReader 初始化AllReader
func (this *Client) initAllReader() {
	var f ios.FReader
	if this.Event.OnReadFrom != nil {
		f = ios.FReadFunc(this.Event.OnReadFrom)
	}
	switch v := this.AllReader.(type) {
	case *ios.MoreRead:
		if v != nil {
			v.Reset(this.buf, f)
			return
		}
	}
	this.AllReader = ios.NewAllReader(this.buf, f)
}

func (this *Client) SetReadWriteCloser(k string, r ios.ReadWriteCloser) {
	this.key = k
	this.r = r

	//设置缓存区4KB,针对io.Reader有效,能大幅度提升性能
	//这个是缓存区,和实际读取的buffer不一样,固有2个内存的申明,
	//io经常什么释放,需要注意内存的释放问题
	//所以固定了size为4kb,方便内存的复用,减少(频繁重连)内存泄漏情况
	switch v := r.(type) {
	case io.Reader:
		buf := bufferReadePool.Get()
		buf.Reset(v)
		this.buf = buf
	default:
		this.buf = r
	}

	//需要先初始化，方便OnConnect的数据读取,run的时候还会声明一次最新(用户设置过)的读取函数
	//转换为FreeFromReader,附带内存释放的FromReader
	//Event中的内存由用户自行控制,如果未配置(nil),则由全局pool控制生成
	this.initAllReader()
	moreWrite := ios.NewMoreWrite(r)
	this.MoreWriter = moreWrite
	this.Info.DialTime = time.Now()
	//this.options = op

	//Runner现在作为全局的生命周期,Closer控制单次的生命周期
	//this.Runner = safe.NewRunnerWithContext(this.Ctx, this.run)

	//写入数据事件
	moreWrite.Option = []ios.WriteOption{
		func(p []byte) (_ []byte, err error) {
			this.Logger.Writeln("["+this.GetKey()+"] ", p)
			if this.Event.OnWriteWith != nil {
				p, err = this.Event.OnWriteWith(p)
			}
			this.Info.WriteTime = time.Now()
			this.Info.WriteCount++
			this.Info.WriteBytes += len(p)
			return p, err
		},
	}
	//写入事件
	moreWrite.OnWrite = func(f func() error) error {
		if this.Event != nil && this.Event.OnWrite != nil {
			return this.Event.OnWrite(f)
		}
		return f()
	}

	//重置Closer,非重新申明,节约内存
	this.Closer.Reset()
	this.Closer.SetCloseFunc(func(err error) error {
		//关闭真实实例
		if er := r.Close(); er != nil {
			return er
		}
		//关闭超时机制
		this.timeout.Stop()

		//关闭/断开连接事件
		this.Logger.Errorf("[%s] 断开连接: %s\n", this.GetKey(), err.Error())
		if this.Event.OnDisconnect != nil {
			this.Event.OnDisconnect(this, err)
		}

		//释放内存,读取数据的时候申明了内存,需要释放下,防止内存泄漏
		//释放Reader的内存
		switch v := this.buf.(type) {
		case *ios.BufferReader:
			if v != nil {
				bufferReadePool.Put(v)
			}
			this.buf = nil
		}

		return nil
	})

	//执行选项
	this.SetOption(this.options...)

}

// SetDial 设置连接函数
func (this *Client) SetDial(dial ios.DialFunc) {
	this.dial = dial
}

// Dial 建立连接
func (this *Client) Dial(ctx context.Context) error {

	r, k, err := this._dial(ctx)
	if err != nil {
		this.Closer.Reset()
		this.Closer.CloseWithErr(err)
		return err
	}

	this.SetReadWriteCloser(k, r)

	//打印日志,由op选项控制是否输出和日志等级
	this.Logger.Infof("[%s] 连接服务成功...\n", this.GetKey())

	//触发连接事件
	if this.Event.OnConnected != nil {
		if err := this.Event.OnConnected(this); err != nil {
			this.CloseWithErr(err)
			return err
		}
	}

	//增加连接成功的信号,方便一些逻辑判断
	//当连接成功的时候会发送信号通道(防止代码错误最多20个),则监听能收到连接成功的信号
	//当连接关闭的时候,需要重新声明信号,方便下次阻塞
	for i := 0; i < 20; i++ {
		select {
		case this.dialedSign <- struct{}{}:
		default:
			break
		}
	}

	return nil
}

func (this *Client) _dial(ctx context.Context) (ios.ReadWriteCloser, string, error) {
	if this.dial == nil {
		return nil, "", errors.New("dial function is nil")
	}
	//首次连接
	r, k, err := this.dial(ctx)
	if err == nil || !this.redial {
		return r, k, err
	}

	var getInterval func(int) (time.Duration, error)
	switch {
	case this.Event != nil && this.Event.OnReconnect != nil:
		getInterval = this.Event.OnReconnect
	default:
		getInterval = defaultReconnect
	}

	//尝试重连
	for i := 1; this.redial && err != nil; i++ {
		t, er := getInterval(i)
		if er != nil {
			return nil, "", er
		}
		if this.GetKey() != "" {
			k = this.GetKey()
		}
		this.Logger.Errorf("[%s] %v,等待%d秒重试\n", k, common.DealErr(err), t/time.Second)
		select {
		case <-ctx.Done():
			return nil, "", ctx.Err()
		case <-time.After(t):
			r, k, err = this.dial(ctx)
		}
	}
	return r, k, err
}

// SetReadTimeout 设置读取超时,即距离上次读取数据时间超过该设置值,则会关闭连接,0表示不超时,todo 逻辑整理的更清晰点
func (this *Client) SetReadTimeout(ctx context.Context, timeout time.Duration) *Client {
	this.timeout.SetFunc(func(ctx context.Context) error {
		if timeout <= 0 {
			return nil
		}
		timer := time.NewTimer(timeout)
		defer timer.Stop()
		for {
			timer.Reset(timeout)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timer.C:
				this.CloseWithErr(ios.ErrReadTimeout)
				return ios.ErrReadTimeout
			}
		}
	})

	//不用判断客户端是否已经运行,可能还没开始执行,可能还未调用Run
	//if this.timeout.Running() {
	this.timeout.Restart(ctx)
	//}

	return this
}

// SetOption 设置选项,立马执行
func (this *Client) SetOption(op ...Option) *Client {
	for _, fn := range op {
		fn(this)
	}
	return this
}

// GetKey 获取标识
func (this *Client) GetKey() string {
	return this.key
}

func (this *Client) SetKey(key string) *Client {
	oldKey := this.key
	this.key = key
	if this.key != oldKey {
		this.Logger.Infof("[%s] 修改标识为 [%s]\n", oldKey, this.key)
		if this.Event.OnKeyChange != nil {
			this.Event.OnKeyChange(this, oldKey)
		}
	}
	return this
}

func (this *Client) Timer(t time.Duration, f func(c *Client)) {
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

// dealErr 自定义错误信息,例如把英文信息改中文
func (this *Client) dealErr(err error) error {
	if err == nil {
		return nil
	}
	if this.Event != nil && this.Event.OnDealErr != nil {
		return this.Event.OnDealErr(this, err)
	}
	return err
}

// GoTimerWriter 定时写入,容易忘记使用协程,然后阻塞,索性直接用协程
func (this *Client) GoTimerWriter(t time.Duration, f func(w ios.MoreWriter) error) {
	go this.Timer(t, func(c *Client) {
		select {
		case <-c.Closer.Done():
			return
		default:
			err := f(c)
			err = this.dealErr(err)
			c.CloseWithErr(err)
		}
	})
}

// CloseAll 关闭连接,并不再重试
func (this *Client) CloseAll() error {
	this.SetRedial(false)
	return this.Closer.Close()
}

// SetRedial 设置自动重连,当连接断开时,
// 会进行自动重连,退避重试,直到成功,除非上下文关闭
func (this *Client) SetRedial(b ...bool) *Client {
	this.redial = len(b) == 0 || b[0]
	return this
}

// Redial 断开重连,是否有必要? 因为可以用其他方式实现
func (this *Client) Redial() {
	this.redialSign <- struct{}{}
}

// Dialed 连接成功的信号,每次连接成功都会释放信号(最多20个)
// 需要判断下closer是否关闭,才能保证逻辑更有效
func (this *Client) Dialed() <-chan struct{} {
	return this.dialedSign
}

// Done 这个是客户端生命周期结束的关闭信号,显示申明下,避免Done冲突
func (this *Client) Done() <-chan struct{} {
	return this.Runner2.Done()
}

// run 运行读取数据操作,如果设置了重试,则会自动重连
func (this *Client) run(ctx context.Context) error {
	//判断是否建立了连接,未建立则尝试建立
	if this.r == nil {
		if err := this.Dial(ctx); err != nil {
			return err
		}
	}
	return this._run(ctx)
}

// _run 运行读取数据操作,如果设置了重试,则这个run结束后立马执行run,递归下去,是否会有资源未释放?
func (this *Client) _run(ctx context.Context) (err error) {

	//校验事件函数
	if this.Event == nil {
		this.Event = &Event{}
	}

	//运行的时候,重新加载下OnReadFrom,因为用户的Event是后设置的,固重新加载下
	this.initAllReader()

	//超时机制
	this.timeout.Start(ctx)

	for {
		select {

		case <-ctx.Done():
			//上下文关闭
			return ctx.Err()

		case <-this.Closer.Done():
			//一个连接的生命周期结束
			if !this.redial {
				//如果未设置重试,则直接返回错误
				return this.Closer.Err()
			}
			//设置了重连,并且已经运行,其他都关闭
			//这里连接的错误只会出现在上下文关闭的情况
			if err := this.Dial(ctx); err != nil {
				return err
			}
			return this._run(ctx)

		case <-this.redialSign:

			//先关闭老连接
			this.CloseWithErr(errors.New("手动重连"))
			//尝试建立连接,不需要重试,连接失败后会进行下一个循环
			//下个循环会走正常的断开是否重连逻辑,设置重连会一直重试,否则退出执行
			this.Dial(ctx)

		default:

		}

		//读取数据,目前支持3种类型,Reader, AReader, MReader
		//如果是AReader,MReader,说明是分包分好的数据,则直接读取即可
		//如果是Reader,则数据还处于粘包状态,需要调用时间OnReadBuffer,来进行读取
		ack, err := this.ReadAck()
		if err != nil {
			//自定义错误信息,例如把英文信息改中文
			err = this.dealErr(err)
			this.CloseWithErr(err)
			//交给closer进行处理接下来的逻辑,固这里不使用return
			//例如重新连接等操作,这样只用写一个地方,简化代码
			continue
		}

		//数据读取成功,更新时间等信息
		this.Info.ReadTime = time.Now()
		this.Info.ReadCount++
		this.Info.ReadBytes += len(ack.Payload())

		//处理数据,使用事件OnDealMessage处理数据,
		//如果未实现,则不处理数据,并确认消息
		this.Logger.Readln("["+this.GetKey()+"] ", ack.Payload())
		if this.Event.OnDealMessage != nil {
			this.Event.OnDealMessage(this, ack)
			continue
		}
		//未设置处理事件,则直接确认
		ack.Ack()

	}

}
