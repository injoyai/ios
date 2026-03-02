package server

import (
	"github.com/injoyai/ios/v2/client"
)

type Event struct {
	onOpen         func(s *Server)                              //服务开启事件
	onClose        func(s *Server, err error)                   //服务关闭事件
	onConnected    func(s *Server, c *client.Client) error      //客户端连接事件
	onDisConnected func(s *Server, c *client.Client, err error) //客户端断开连接事件
}

func (this *Event) OnOpen(f func(s *Server)) {
	this.onOpen = f
}

func (this *Event) OnClose(f func(s *Server)) {
	this.onOpen = f
}

func (this *Event) OnConnected(f func(s *Server, c *client.Client) error) {
	this.onConnected = f
}

func (this *Event) OnDisConnected(f func(s *Server, c *client.Client, err error)) {
	this.onDisConnected = f
}

/*



 */

func WithClient(op ...client.Option) Option {
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
