package client

import (
	"context"
	"github.com/injoyai/conv"
	"github.com/injoyai/ios"
	"github.com/injoyai/ios/module/common"
	"io"
	"sync"
	"time"
)

type Frame interface {
	ReadFrom(r io.Reader) ([]byte, error) //读取数据事件,当类型是io.Reader才会触发
	WriteWith(bs []byte) ([]byte, error)  //写入消息事件
}

type Event struct {
	OnConnected   func(c *Client) error                                                                        //连接事件
	OnReconnect   func(ctx context.Context, c *Client, dial ios.DialFunc) (ios.ReadWriteCloser, string, error) //重新连接事件
	OnDisconnect  func(c *Client, err error)                                                                   //断开连接事件
	OnReadFrom    func(r io.Reader) ([]byte, error)                                                            //读取数据事件,当类型是io.Reader才会触发
	OnDealMessage func(c *Client, msg ios.Acker)                                                               //处理消息事件
	OnWriteWith   func(bs []byte) ([]byte, error)                                                              //写入消息数据事件,例如封装数据格式
	OnWrite       func(f func() error) error                                                                   //写入消息事件,例如并发安全,错误重试
	OnKeyChange   func(c *Client, oldKey string)                                                               //修改标识事件
	OnDealErr     func(c *Client, err error) error                                                             //修改错误信息事件,例翻译成中文
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

// NewReconnectInterval 按一定时间间隔进行重连
func NewReconnectInterval(t time.Duration) func(ctx context.Context, c *Client, dial ios.DialFunc) (ios.ReadWriteCloser, string, error) {
	return func(ctx context.Context, c *Client, dial ios.DialFunc) (ios.ReadWriteCloser, string, error) {
		r, k, err := dial(ctx)
		if err == nil {
			return r, k, nil
		}
		for {
			select {
			case <-ctx.Done():
				return nil, "", ctx.Err()
			case <-time.After(t):
				r, k, err = dial(ctx)
				if err == nil {
					return r, k, nil
				}
				if c.GetKey() != "" {
					k = c.GetKey()
				}
				c.Logger.Errorf("[%s] %v,等待%d秒重试\n", k, common.DealErr(err), t/time.Second)
			}
		}
	}
}

var (
	// defaultReconnect 默认重连机制
	defaultReconnect = NewReconnectRetreat(time.Second*2, time.Second*32, 2)
)

// NewReconnectRetreat 退避重试
func NewReconnectRetreat(start, max time.Duration, multi uint8) func(ctx context.Context, c *Client, dial ios.DialFunc) (ios.ReadWriteCloser, string, error) {
	if start < 0 {
		start = time.Second * 2
	}
	if max < start {
		max = start
	}
	if multi == 0 {
		multi = 2
	}
	return func(ctx context.Context, c *Client, dial ios.DialFunc) (ios.ReadWriteCloser, string, error) {
		wait := time.Second * 0
		for i := 0; ; i++ {
			select {
			case <-ctx.Done():
				return nil, "", ctx.Err()
			case <-time.After(wait):
				r, k, err := dial(ctx)
				if err == nil {
					return r, k, nil
				}
				if wait < start {
					wait = start
				} else if wait < max {
					wait *= time.Duration(multi)
				}
				if wait >= max {
					wait = max
				}
				if c.GetKey() != "" {
					k = c.GetKey()
				}
				c.Logger.Errorf("[%s] %v,等待%d秒重试\n", k, common.DealErr(err), wait/time.Second)
			}
		}
	}
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
