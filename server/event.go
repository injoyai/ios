package server

import (
	"github.com/injoyai/ios/client"
)

type Event struct {
	OnOpen      func(s *Server)                         //服务开启事件
	OnClose     func(s *Server, err error)              //服务关闭事件
	OnConnected func(s *Server, c *client.Client) error //客户端连接事件
}
