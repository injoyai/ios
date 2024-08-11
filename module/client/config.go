package client

import (
	"context"
	"github.com/injoyai/ios"
)

type Event struct {
	OnConnect      func(c *Client) error                             //连接事件
	OnReadBuffer   func(r ios.Reader, buf []byte) (ios.Acker, error) //读取数据事件
	OnDealMessage  func(c *Client, message ios.Acker)                //处理消息事件
	OnWriteMessage func(bs []byte) ([]byte, error)                   //写入消息事件
	OnDisconnect   func(ctx context.Context, c *Client, err error)   //断开连接事件
}
