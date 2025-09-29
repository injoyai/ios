package client

import (
	"github.com/injoyai/conv"
	"github.com/injoyai/ios"
	"io"
	"math"
	"sync"
	"time"
)

type Frame interface {
	ReadFrom(r io.Reader) ([]byte, error) //读取数据事件,当类型是io.Reader才会触发
	WriteWith(bs []byte) ([]byte, error)  //写入消息事件
}

type Event struct {
	OnConnected   func(c *Client) error              //连接事件
	OnReconnect   func(i int) (time.Duration, error) //重连事件
	OnDisconnect  func(c *Client, err error)         //断开连接事件
	OnReadFrom    func(r io.Reader) ([]byte, error)  //读取数据事件,当类型是io.Reader才会触发
	OnDealMessage func(c *Client, msg ios.Acker)     //处理消息事件
	OnWriteWith   func(bs []byte) ([]byte, error)    //写入消息数据事件,例如封装数据格式
	OnWrite       func(f func() error) error         //写入消息事件,例如并发安全,错误重试
	OnKeyChange   func(c *Client, oldKey string)     //修改标识事件
	OnDealErr     func(c *Client, err error) error   //修改错误信息事件,例翻译成中文
}

func (this *Event) WithFrame(f Frame) {
	this.OnReadFrom = f.ReadFrom
	this.OnWriteWith = f.WriteWith
}

type Info struct {
	CreateTime time.Time //创建时间,对象创建时间,重连不会改变
	DialTime   time.Time //连接时间,每次重连会改变
	ReadTime   time.Time //本次连接,最后读取到数据的时间
	ReadCount  int       //本次连接,读取数据次数
	ReadBytes  int       //本次连接,读取数据字节
	WriteTime  time.Time //本次连接,最后写入数据时间
	WriteCount int       //本次连接,写入数据次数
	WriteBytes int       //本次连接,写入数据字节
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

var (
	// defaultReconnectInterval 默认重连时间间隔
	defaultReconnect = NewReconnectRetreat(time.Second*2, time.Second*32, 2)
)

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
		if _, err := w.Write(msg.Payload()); err == nil {
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
