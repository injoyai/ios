package client

import (
	"bufio"
	"context"
	"errors"
	"github.com/injoyai/ios"
	"time"
)

type Handler func(ctx context.Context) (ios.ReadeWriteCloser, string, error)

type Config struct {
	Dial    Handler  //连接函数
	Logger  Logger   //日志
	Options []Option //选项
	Event            //事件
}

type Event struct {
	OnConnect      func(c *Client) error                           //连接事件
	OnReadBuffer   func(buf *bufio.Reader) ([]byte, error)         //读取数据事件
	OnDealMessage  func(c *Client, msg Message)                    //处理消息事件
	OnWriteMessage func(bs []byte) ([]byte, error)                 //写入消息事件
	OnDisconnect   func(ctx context.Context, c *Client, err error) //断开连接事件
}

func (this *Config) defaultRedial(ctx context.Context, h Handler) (ios.ReadeWriteCloser, string, error) {
	if h == nil {
		return nil, "", errors.New("handler is nil")
	}
	wait := time.Second * 0
	for i := 0; ; i++ {
		select {
		case <-ctx.Done():
			return nil, "", ctx.Err()
		case <-time.After(wait):
			c, s, err := h(ctx)
			if err == nil {
				return c, s, nil
			}
			if wait < time.Second {
				wait = time.Second
			} else if wait <= time.Second*16 {
				wait *= 2
			}
			this.Logger.Errorf("%v,等待%d秒重试\n", err, wait/time.Second)
		}
	}
}
