package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/injoyai/base/maps"
	"github.com/injoyai/base/safe"
	"github.com/injoyai/ios/v2"
	"github.com/injoyai/ios/v2/module/common"
)

func Run(dial ios.DialFunc, op ...Option) error {
	return RunContext(context.Background(), dial, op...)
}

func RunContext(ctx context.Context, dial ios.DialFunc, op ...Option) error {
	c, err := Dial(dial, op...)
	if err != nil {
		return err
	}
	return c.Run(ctx)
}

func Redial(dial ios.DialFunc, op ...Option) *Client {
	return RedialContext(context.Background(), dial, op...)
}

func RedialContext(ctx context.Context, dial ios.DialFunc, op ...Option) *Client {
	c, _ := DialContext(ctx, dial, func(c *Client) {
		c.SetOption(op...)
		c.SetRedial()
	})
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
	c := &Client{
		Logger:     common.NewLogger(),
		Info:       newInfo(),
		event:      newEvent(),
		Closer:     safe.NewCloserErr(errors.New("等待连接")),
		Runner2:    safe.NewRunner2(nil),
		redialSign: make(chan struct{}),
		dial:       dial,
	}
	c.Runner2.SetFunc(c.run)
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
	*event

	//安全关闭,单次的生命周期,每次重连都会重新声明
	*safe.Closer

	//运行,全局的生命周期,包括重试
	*safe.Runner2

	//日志管理,默认使用common.NewLogger
	Logger common.Logger

	/*
		internal 内部字段
	*/

	//IO实例,原始数据
	r ios.ReadWriteCloser

	//基于r,带缓存的reader
	//目前支持ios.AReader,ios.MReader,io.Reader
	buf ios.Reader

	//全局自定义标识,表明客户端的身份
	//默认使用的是客户端的IP:PORT
	//例如,通过解析注册信息后,可以使用解析的IMEI等信息作为key
	key string

	//标签,用于自定义记录连接的一些信息
	//例如,客户端的ICCID,IMEI等
	tag     *maps.Safe
	tagOnce sync.Once

	//超时机制,监听客户端的读取和写入数据,维持不超时
	timeout time.Duration

	//全局重连信号,未设置自动重连也可以手动重连,
	//向这个通道发送一个信号,则客户端会进行断开重连
	redialSign chan struct{}

	//是否自动重连
	//当连接断开的时候,自动重连
	//还是同一个客户端
	redial bool

	//缓存连接函数,重连的时候使用
	dial ios.DialFunc
}

// Origin 获取原始连接,
// 尽量不要直接进行读取,
// 因为内部封装了一层buffer,可能会造成数据混乱
func (this *Client) Origin() ios.ReadWriteCloser {
	return this.r
}

func (this *Client) SetReadWriteCloser(key string, r ios.ReadWriteCloser) {
	this.key = key
	this.r = r

	//设置缓存区4KB,针对io.Reader有效,能大幅度提升性能
	//这个是缓存区,和实际读取的buffer不一样,固有2个内存的申明,
	//io经常什么释放,需要注意内存的释放问题
	//所以固定了size为4kb,方便内存的复用,减少(频繁重连)内存泄漏情况
	switch v := r.(type) {
	case io.Reader:
		buf := DefaultPool.Get()
		buf.Reset(v)
		this.buf = buf
	default:
		this.buf = r
	}

	//需要先初始化，方便OnConnect的数据读取,run的时候还会声明一次最新(用户设置过)的读取函数
	//转换为FreeFromReader,附带内存释放的FromReader
	//Event中的内存由用户自行控制,如果未配置(nil),则由全局pool控制生成
	if this.event.onReadFrom == nil {
		this.AllReader = ios.NewAllReader(this.buf, nil)
	} else {
		this.AllReader = ios.NewAllReader(this.buf, ios.FReadFunc(this.event.onReadFrom))
	}
	this.Info.DialTime = time.Now()

	//设置多种方式写入
	this.MoreWriter = ios.NewMoreWrite(
		r,
		func(p []byte, write func(p []byte) error) (err error) {
			this.Logger.Writeln("["+this.Key()+"] ", p)
			for _, f := range this.event.onWriteWith {
				p, err = f(p)
			}
			this.Info.WriteTime = time.Now()
			this.Info.WriteCount++
			this.Info.WriteBytes += int64(len(p))
			if this.event.onWrite != nil {
				return this.event.onWrite(func() error { return write(p) })
			}
			return write(p)
		},
	)

	//重置Closer,非重新申明,节约内存
	this.Closer.Reset()
	this.Closer.SetCloseFunc(func(err error) error {

		//关闭真实实例
		if er := r.Close(); er != nil {
			return er
		}

		//关闭/断开连接事件
		this.Logger.Errorf("[%s] 断开连接: %s\n", this.Key(), err.Error())
		for _, f := range this.event.onDisconnect {
			if f != nil {
				f(this, err)
			}
		}

		//释放内存,读取数据的时候申明了内存,需要释放下,防止内存泄漏
		//释放Reader的内存
		switch v := this.buf.(type) {
		case *ios.BufferReader:
			if v != nil {
				DefaultPool.Put(v)
			}
			this.buf = nil
		}

		return nil
	})

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
	this.Logger.Infof("[%s] 连接服务成功...\n", this.Key())

	//触发连接事件
	if this.event.onConnected != nil {
		if err := this.event.onConnected(this); err != nil {
			this.CloseWithErr(err)
			return err
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

	var getInterval = defaultReconnect
	if this.event.onReconnect != nil {
		getInterval = this.event.onReconnect
	}

	//尝试重连
	for i := 1; this.redial && err != nil; i++ {
		t, er := getInterval(i)
		if er != nil {
			return nil, "", er
		}
		if this.Key() != "" {
			k = this.Key()
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

// SetReadTimeout 设置读取超时
func (this *Client) SetReadTimeout(timeout time.Duration) *Client {
	this.timeout = timeout
	return this
}

// SetOption 设置选项,立马执行
func (this *Client) SetOption(op ...Option) *Client {
	for _, fn := range op {
		fn(this)
	}
	return this
}

// Key 获取标识
func (this *Client) Key() string {
	return this.key
}

// SetKey 设置标识
func (this *Client) SetKey(key string) *Client {
	oldKey := this.key
	this.key = key
	if this.key != oldKey {
		this.Logger.Infof("[%s] 修改标识为 [%s]\n", oldKey, this.key)
		for _, f := range this.event.onKeyChange {
			f(this, oldKey)
		}
	}
	return this
}

// Tag 标签
func (this *Client) Tag() *maps.Safe {
	this.tagOnce.Do(func() { this.tag = maps.NewSafe() })
	return this.tag
}

// Timer 定时任务
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

// GoAfter 协程延迟执行
func (this *Client) GoAfter(t time.Duration, f func(c *Client)) {
	go func() {
		select {
		case <-this.Closer.Done():
			return
		case <-time.After(t):
			f(this)
		}
	}()
}

// Exit 退出,并不再重试
func (this *Client) Exit() error {
	this.SetRedial(false)
	return this.Closer.Close()
}

// SetRedial 设置自动重连,当连接断开时,
// 会进行自动重连,退避重试,直到成功,除非上下文关闭
func (this *Client) SetRedial(b ...bool) *Client {
	this.redial = len(b) == 0 || b[0]
	return this
}

// Redial 手动断开重连
func (this *Client) Redial() {
	select {
	case this.redialSign <- struct{}{}:
	default:
	}
}

// Done 这个是客户端生命周期结束的关闭信号,显示申明下,避免Done冲突
func (this *Client) Done() <-chan struct{} {
	return this.Runner2.Done()
}

// dealErr 自定义错误信息,例如把英文信息改中文
func (this *Client) dealErr(err error) error {
	if err == nil {
		return nil
	}
	if this.event.onDealErr == nil {
		return err
	}
	return this.event.onDealErr(this, err)
}

// runTimeout 执行超时机制
func (this *Client) runTimeout(ctx context.Context) error {
	if this.timeout <= 0 {
		return nil
	}
	timer := time.NewTimer(this.timeout)
	defer timer.Stop()
	for {
		if this.timeout <= 0 {
			return nil
		}
		timer.Reset(this.timeout)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			this.CloseWithErr(ios.ErrReadTimeout)
			return ios.ErrReadTimeout
		}
	}
}

// run 运行读取数据操作,如果设置了重试,则会自动重连
func (this *Client) run(ctx context.Context) error {
	for {

		//判断是否建立了连接,未建立则尝试建立
		if this.r == nil {
			if err := this.Dial(ctx); err != nil {
				return err
			}
		}

		//运行
		redial, err := this._run(ctx)

		//判断是否需要重连,不重连返回错误
		if !this.redial && !redial {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
			//避免无限制重连
		}

		this.r = nil

	}
}

// _run 运行读取数据操作
func (this *Client) _run(ctx context.Context) (redial bool, err error) {

	defer func() {
		if e := recover(); e != nil {
			redial = false
			err = fmt.Errorf("%v", e)
		}
		err = this.dealErr(err)
		this.CloseWithErr(err)
	}()

	//判断是否能设置读超时
	deadliner, isDeadliner := this.r.(ios.SetReadDeadliner)
	if !isDeadliner {
		go this.runTimeout(ctx)
	}

	for {

		//设置读超时,如果能设置的话
		if isDeadliner && this.timeout > 0 {
			err = deadliner.SetReadDeadline(time.Now().Add(this.timeout))
			if err != nil {
				return
			}
		}

		select {

		case <-ctx.Done():
			//上下文关闭
			return false, ctx.Err()

		case <-this.Closer.Done():
			//一个连接的生命周期结束
			return false, this.Closer.Err()

		case <-this.redialSign:
			//手动关闭连接,然后重连1次
			return true, errors.New("手动重连")

		default:

		}

		//读取数据,目前支持3种类型,Reader, AReader, MReader
		//如果是AReader,MReader,说明是分包分好的数据,则直接读取即可
		//如果是Reader,则数据还处于粘包状态,需要调用时间OnReadFrom,来进行读取
		ack, err := this.ReadAck()
		if err != nil {
			return false, err
		}

		//数据读取成功,更新时间等信息
		this.Info.ReadTime = time.Now()
		this.Info.ReadCount++
		this.Info.ReadBytes += int64(len(ack.Bytes()))

		//处理数据,使用事件OnDealMessage处理数据,
		//如果未实现,则不处理数据,并确认消息
		this.Logger.Readln("["+this.Key()+"] ", ack.Bytes())
		for _, f := range this.onDealMessage {
			f(this, ack)
		}
		if len(this.onDealMessage) > 0 {
			continue
		}
		//未设置处理事件,则直接确认
		ack.Ack()

	}

}
