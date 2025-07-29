package client

import (
	"github.com/injoyai/ios"
	"io"
)

type Option func(c *Client)

// WithRedial 重连
func WithRedial(b ...bool) Option {
	return func(c *Client) {
		c.SetRedial(b...)
	}
}

// WithDebug 调试
func WithDebug(b ...bool) Option {
	return func(c *Client) {
		c.Logger.Debug(b...)
	}
}

// WithLevel 设置日志等级
func WithLevel(l int) Option {
	return func(c *Client) {
		c.Logger.SetLevel(l)
	}
}

// WithHEX 日志以hex输出
func WithHEX() Option {
	return func(c *Client) {
		c.Logger.WithHEX()
	}
}

// WithUTF8 日志以utf8输出
func WithUTF8() Option {
	return func(c *Client) {
		c.Logger.WithUTF8()
	}
}

// WithDealMessage 处理消息事件
func WithDealMessage(f func(c *Client, msg ios.Acker)) Option {
	return func(c *Client) {
		c.Event.OnDealMessage = f
	}
}

// WithReadFrom 读取数据事件
func WithReadFrom(f func(r io.Reader) ([]byte, error)) Option {
	return func(c *Client) {
		c.Event.OnReadFrom = f
	}
}

// WithWriteWith 写入数据事件
func WithWriteWith(f func(bs []byte) ([]byte, error)) Option {
	return func(c *Client) {
		c.Event.OnWriteWith = f
	}
}

// WithFrame 设置Frame
func WithFrame(f Frame) Option {
	return func(c *Client) {
		c.Event.WithFrame(f)
	}
}

// WithConnect 建立连接事件
func WithConnect(f func(c *Client) error) Option {
	return func(c *Client) {
		c.Event.OnConnected = f
	}
}

// WithDisconnect 断开连接事件
func WithDisconnect(f func(c *Client, err error)) Option {
	return func(c *Client) {
		c.Event.OnDisconnect = f
	}
}
