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
