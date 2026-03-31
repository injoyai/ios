package client

import (
	"bufio"
	"bytes"
	"io"
	"math"
	"sync"
	"time"

	"github.com/injoyai/conv"
	"github.com/injoyai/ios/v2"
)

func newEvent() *event {
	return &event{
		onReconnect: defaultReconnect,
		onDealErr:   defaultDealErr,
	}
}

type Frame interface {
	ReadFrom(r *bufio.Reader) ([]byte, error) //读取数据事件,当类型是io.Reader才会触发
	WriteWith(bs []byte) ([]byte, error)      //写入消息事件
}

type event struct {
	onConnected   []Option                           //连接事件
	onReconnect   func(i int) (time.Duration, error) //重连事件,i是重连次数,1开始
	onDisconnect  []func(c *Client, err error)       //断开连接事件
	onReadFrom    ios.FReadFunc                      // func(r *bufio.Reader) ([]byte, error) //读取数据事件,当类型是io.Reader才会触发
	onDealMessage []func(c *Client, msg ios.Acker)   //处理消息事件
	onWriteWith   []func(bs []byte) ([]byte, error)  //写入消息数据事件,例如封装数据格式
	onWrite       func(f func() error) error         //写入消息事件,例如并发安全,错误重试
	onKeyChange   []func(c *Client, oldKey string)   //修改标识事件
	onDealErr     func(c *Client, err error) error   //修改错误信息事件,例翻译成中文
}

func (this *event) OnConnected(f ...Option) {
	this.onConnected = append(this.onConnected, f...)
}

func (this *event) DoConnected(c *Client) {
	c.SetOption(this.onConnected...)
}

func (this *event) OnReconnect(f func(i int) (time.Duration, error)) {
	if f != nil {
		this.onReconnect = f
	}
}

func (this *event) OnReconnectInterval(t time.Duration) {
	this.OnReconnect(func(i int) (time.Duration, error) { return t, nil })
}

func (this *event) OnDisconnect(f func(c *Client, err error)) {
	if f != nil {
		this.onDisconnect = append(this.onDisconnect, f)
	}
}

func (this *event) OnReadFrom(f ios.FReadFunc) {
	if f != nil {
		this.onReadFrom = f
	}
}

func (this *event) OnReadWithSplit(delim byte, trim ...bool) {
	this.OnReadFrom(func(r *bufio.Reader) ([]byte, error) {
		line, err := r.ReadBytes(delim)
		if len(trim) > 0 && trim[0] {
			line = bytes.Trim(line, string(delim))
		}
		return line, err
	})
}

func (this *event) OnDealMessage(f func(c *Client, msg ios.Acker)) {
	if f != nil {
		this.onDealMessage = append(this.onDealMessage, f)
	}
}

func (this *event) OnWriteWith(f func(bs []byte) ([]byte, error)) {
	if f != nil {
		this.onWriteWith = append(this.onWriteWith, f)
	}
}

func (this *event) OnWriteWithPrefix(prefix []byte) {
	this.OnWriteWith(func(bs []byte) ([]byte, error) {
		return append(prefix, bs...), nil
	})
}

func (this *event) OnWriteWithSuffix(suffix []byte) {
	this.OnWriteWith(func(bs []byte) ([]byte, error) {
		return append(bs, suffix...), nil
	})
}

func (this *event) OnWrite(f func(f func() error) error) {
	if f != nil {
		this.onWrite = f
	}
}

func (this *event) OnKeyChange(f func(c *Client, oldKey string)) {
	if f != nil {
		this.onKeyChange = append(this.onKeyChange, f)
	}
}

func (this *event) OnDealErr(f func(c *Client, err error) error) {
	if f == nil {
		f = func(c *Client, err error) error { return err }
	}
	this.onDealErr = f
}

func (this *event) WithFrameSplit(delim byte, trim ...bool) {
	this.OnReadWithSplit(delim, trim...)
	this.OnWriteWithSuffix([]byte{delim})
}

func (this *event) WithFrame(f Frame) {
	this.OnReadFrom(f.ReadFrom)
	this.OnWriteWith(f.WriteWith)
}

func newInfo() Info {
	return Info{CreateTime: time.Now()}
}

type Info struct {
	CreateTime time.Time //创建时间,对象创建时间,重连不会改变
	DialTime   time.Time //连接时间,每次重连会改变
	ReadTime   time.Time //本次连接,最后读取到数据的时间
	ReadCount  int64     //本次连接,读取数据次数
	ReadBytes  int64     //本次连接,读取数据字节
	WriteTime  time.Time //本次连接,最后写入数据时间
	WriteCount int64     //本次连接,写入数据次数
	WriteBytes int64     //本次连接,写入数据字节
}

// NewWriteSafe 写入并发安全,例如websocket不能并发写入
func NewWriteSafe() func(f func() error) error {
	mu := sync.Mutex{}
	return func(f func() error) error {
		mu.Lock()
		defer mu.Unlock()
		return f()
	}
}

// NewWriteRetry 写入错误重试
func NewWriteRetry(retry int, interval ...time.Duration) func(f func() error) error {
	after := conv.Default(0, interval...)
	return func(f func() error) (err error) {
		for i := 0; i <= retry; i++ {
			if err = f(); err == nil {
				break
			}
			<-time.After(after)
		}
		return
	}
}

// NewReconnectRetreat 退避重试
func NewReconnectRetreat(min, max time.Duration, base int) func(i int) (time.Duration, error) {
	return func(i int) (time.Duration, error) {
		n := math.Pow(float64(base), float64(i))
		t := time.Second * time.Duration(n)
		return conv.Range(t, min, max), nil
	}
}

// NewReconnectInterval 按一定时间间隔进行重连
func NewReconnectInterval(t time.Duration) func(i int) (time.Duration, error) {
	return func(i int) (time.Duration, error) { return t, nil }
}

// NewDealMessageWithChan 把数据写入到chan中
func NewDealMessageWithChan(ch chan ios.Acker) func(c *Client, msg ios.Acker) {
	return func(c *Client, msg ios.Acker) {
		ch <- msg
	}
}

// NewDealMessageWithWriter 把数据写入到io.Writer中
func NewDealMessageWithWriter(w io.Writer) func(c *Client, msg ios.Acker) {
	return func(c *Client, msg ios.Acker) {
		if _, err := w.Write(msg.Bytes()); err == nil {
			msg.Ack()
		}
	}
}

// NewDisconnectAfter 断开连接等待
func NewDisconnectAfter(t time.Duration) func(c *Client, err error) error {
	return func(c *Client, err error) error {
		<-time.After(t)
		return nil
	}
}
