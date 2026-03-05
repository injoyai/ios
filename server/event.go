package server

import (
	"github.com/injoyai/ios/v2/client"
)

type Event struct {
	onOpen        func(s *Server)            //服务开启事件
	onClose       func(s *Server, err error) //服务关闭事件
	clientOptions []client.Option            //客户端选项
}

func (this *Event) OnOpen(f func(s *Server)) {
	this.onOpen = f
}

func (this *Event) OnClose(f func(s *Server, err error)) {
	this.onClose = f
}

func (this *Event) OnClient(op ...client.Option) {
	this.clientOptions = append(this.clientOptions, op...)
}

func (this *Event) OnClientConnected(op ...client.Option) {
	this.clientOptions = append(this.clientOptions, func(c *client.Client) {
		c.OnConnected(op...)
	})
}

/*



 */

func WithClientOptions(op ...client.Option) Option {
	return func(s *Server) {
		s.OnClient(op...)
	}
}

func WithLoggerLevel(level int) Option {
	return func(s *Server) {
		s.Logger.SetLevel(level)
	}
}

func WithLoggerEnable(enable ...bool) Option {
	return func(s *Server) {
		s.Logger.Enable(enable...)
	}
}

func WithLoggerDisable() Option {
	return func(s *Server) {
		s.Logger.Enable(false)
	}
}
