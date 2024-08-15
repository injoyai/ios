package client

import (
	"context"
	"github.com/injoyai/ios"
	"io"
	"time"
)

type Event struct {
	OnConnected    func(c *Client) error                                                             //连接事件
	OnReconnect    func(ctx context.Context, dial ios.DialFunc) (ios.ReadWriteCloser, string, error) //必须连接事件
	OnReadBuffer   func(r io.Reader) ([]byte, error)                                                 //读取数据事件,当类型是io.Reader才会触发
	OnDealMessage  func(c *Client, message ios.Acker)                                                //处理消息事件
	OnWriteMessage func(bs []byte) ([]byte, error)                                                   //写入消息事件
	OnDisconnect   func(c *Client, err error)                                                        //断开连接事件
	OnKeyChange    func(c *Client, oldKey string)                                                    //修改标识事件
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

// WithReconnectInterval 按一定时间间隔进行重连
func WithReconnectInterval(t time.Duration) func(ctx context.Context, dial ios.DialFunc) (ios.ReadWriteCloser, string, error) {
	return func(ctx context.Context, dial ios.DialFunc) (ios.ReadWriteCloser, string, error) {
		for {
			select {
			case <-ctx.Done():
				return nil, "", ctx.Err()
			case <-time.After(t):
				r, k, err := dial(ctx)
				if err == nil {
					return r, k, nil
				}
			}
		}
	}
}

// WithReconnectRetreat 退避重试
func WithReconnectRetreat(start, max time.Duration, multi uint8) func(ctx context.Context, dial ios.DialFunc) (ios.ReadWriteCloser, string, error) {
	if start < 0 {
		start = time.Second * 2
	}
	if max < start {
		max = start
	}
	if multi == 0 {
		multi = 2
	}
	return func(ctx context.Context, dial ios.DialFunc) (ios.ReadWriteCloser, string, error) {
		wait := time.Second * 0
		for i := 0; ; i++ {
			select {
			case <-ctx.Done():
				return nil, "", ctx.Err()
			case <-time.After(wait):
				c, k, err := dial(ctx)
				if err == nil {
					return c, k, nil
				}
				if wait < start {
					wait = start
				} else if wait < max {
					wait *= time.Duration(multi)
				}
				if wait >= max {
					wait = max
				}
			}
		}
	}
}
