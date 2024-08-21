package server

import "github.com/injoyai/ios/module/client"

type Event struct {
	OnListened     func(s *Server)                         //服务监听事件
	OnConnected    func(s *Server, c *client.Client) error //客户端连接事件
	OnKeyChange    func(c *client.Client, oldKey string)
	OnDisconnect   func(c *client.Client, err error)
	OnWriteMessage func(p []byte) ([]byte, error)
}
