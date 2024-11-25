package client

import (
	"github.com/injoyai/ios"
	"github.com/injoyai/ios/module/common"
	"io"
	"time"
)

type Frame interface {
	ReadFrom(r io.Reader) ([]byte, error) //读取数据事件,当类型是io.Reader才会触发
	WriteWith(bs []byte) ([]byte, error)  //写入消息事件
}

type Event struct {
	OnConnected   func(c *Client) error                                                   //连接事件
	OnReconnect   func(c *Client, dial ios.DialFunc) (ios.ReadWriteCloser, string, error) //重新连接事件
	OnDisconnect  func(c *Client, err error)                                              //断开连接事件
	OnReadFrom    func(r io.Reader) ([]byte, error)                                       //读取数据事件,当类型是io.Reader才会触发
	OnDealMessage func(c *Client, msg ios.Acker)                                          //处理消息事件
	OnWriteWith   func(bs []byte) ([]byte, error)                                         //写入消息事件
	OnKeyChange   func(c *Client, oldKey string)                                          //修改标识事件
	OnDealErr     func(c *Client, err error) error                                        //修改错误信息事件,例翻译成中文
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

// ReconnectWithInterval 按一定时间间隔进行重连
func ReconnectWithInterval(t time.Duration) func(c *Client, dial ios.DialFunc) (ios.ReadWriteCloser, string, error) {
	return func(c *Client, dial ios.DialFunc) (ios.ReadWriteCloser, string, error) {
		r, k, err := dial(c.Ctx)
		if err == nil {
			return r, k, nil
		}
		for {
			select {
			case <-c.Ctx.Done():
				return nil, "", c.Ctx.Err()
			case <-time.After(t):
				r, k, err := dial(c.Ctx)
				if err == nil {
					return r, k, nil
				}
			}
		}
	}
}

// ReconnectWithRetreat 退避重试
func ReconnectWithRetreat(start, max time.Duration, multi uint8) func(c *Client, dial ios.DialFunc) (ios.ReadWriteCloser, string, error) {
	if start < 0 {
		start = time.Second * 2
	}
	if max < start {
		max = start
	}
	if multi == 0 {
		multi = 2
	}
	return func(c *Client, dial ios.DialFunc) (ios.ReadWriteCloser, string, error) {
		wait := time.Second * 0
		for i := 0; ; i++ {
			select {
			case <-c.Ctx.Done():
				return nil, "", c.Ctx.Err()
			case <-time.After(wait):
				r, k, err := dial(c.Ctx)
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
				c.Logger.Errorf("[%s] %v,等待%d秒重试\n", c.GetKey(), common.DealErr(err), wait/time.Second)
			}
		}
	}
}

// DealMessageWithChan 把数据写入到chan中
func DealMessageWithChan(ch chan ios.Acker) func(c *Client, msg ios.Acker) {
	return func(c *Client, msg ios.Acker) {
		ch <- msg
	}
}

// DealMessageWithWriter 把数据写入到io.Writer中
func DealMessageWithWriter(w io.Writer) func(c *Client, msg ios.Acker) {
	return func(c *Client, msg ios.Acker) {
		if _, err := w.Write(msg.Payload()); err == nil {
			msg.Ack()
		}
	}
}

// DisconnectWithAfter 断开连接等待
func DisconnectWithAfter(t time.Duration) func(c *Client, err error) error {
	return func(c *Client, err error) error {
		<-time.After(t)
		return nil
	}
}
