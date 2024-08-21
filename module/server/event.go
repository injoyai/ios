package server

import "github.com/injoyai/ios/module/client"

type Event struct {
	OnListened  func(s *Server)                         //服务监听事件
	OnConnected func(s *Server, c *client.Client) error //客户端连接事件
	OnClose     func(s *Server, err error)              //服务关闭事件
}
