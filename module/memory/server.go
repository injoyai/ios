package memory

import (
	"fmt"
	"io"
)

func Listen(key string) *Server {
	s, _ := manage.GetOrSetByHandler(key, func() (interface{}, error) {
		return &Server{Ch: make(chan *Client, 1)}, nil
	})
	manage.Set(s, s)
	return s.(*Server)
}

type Server struct {
	Ch chan *Client
}

func (this *Server) Accept() (io.ReadWriteCloser, string, error) {
	c := <-this.Ch
	return c.sIO(), fmt.Sprintf("%p", c), nil
}

func (this *Server) Close() error {
	//同net关闭服务,不影响已连接的客户端
	manage.Del(this)
	return nil
}
