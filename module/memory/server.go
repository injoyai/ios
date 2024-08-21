package memory

import (
	"fmt"
	"github.com/injoyai/ios"
)

func NewListen(key string) func() (ios.Listener, error) {
	return func() (ios.Listener, error) {
		s, _ := manage.GetOrSetByHandler(key, func() (interface{}, error) {
			return &Server{
				key: key,
				Ch:  make(chan *Client, 1),
			}, nil
		})
		manage.Set(s, s)
		return s.(*Server), nil
	}
}

type Server struct {
	key string
	Ch  chan *Client
}

func (this *Server) Addr() string {
	return this.key
}

func (this *Server) Accept() (ios.ReadWriteCloser, string, error) {
	c := <-this.Ch
	return c.sIO(), fmt.Sprintf("%p", c), nil
}

func (this *Server) Close() error {
	//同net关闭服务,不影响已连接的客户端
	manage.Del(this)
	return nil
}
