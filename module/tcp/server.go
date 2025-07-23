package tcp

import (
	"context"
	"fmt"
	"github.com/injoyai/ios"
	"net"
)

var _ ios.Listener = (*Server)(nil)

func NewListen(port int) func() (ios.Listener, error) {
	return func() (ios.Listener, error) {
		/*
			开启端口复用方式,windows:
			config := net.ListenConfig{
						Control: func(network, address string, c syscall.RawConn) error {
							return c.Control(func(fd uintptr) {
								syscall.SetsockoptInt(syscall.Handle(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
							})
						},
					}
			linux:
			net.ListenConfig{
			        Control: func(network, address string, c syscall.RawConn) error {
			            return c.Control(func(fd uintptr) {
			                // 开启 SO_REUSEADDR
			                syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)

			                // 可选：开启 SO_REUSEPORT（Linux 3.9+）
			                syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEPORT, 1)
			            })
			        },
			    }
		*/
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			return nil, err
		}
		return &Server{
			Listener: listener,
		}, nil
	}
}

func NewListenContext(ctx context.Context, port int) ios.ListenFunc {
	return func() (ios.Listener, error) {
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			return nil, err
		}
		//net.ListenConfig的ctx测试起来没效果
		go func() {
			<-ctx.Done()
			listener.Close()
		}()
		return &Server{
			Listener: listener,
		}, nil
	}
}

type Server struct {
	net.Listener
}

func (this *Server) Close() error {
	return this.Listener.Close()
}

func (this *Server) Accept() (ios.ReadWriteCloser, string, error) {
	c, err := this.Listener.Accept()
	if err != nil {
		return nil, "", err
	}
	return c, c.RemoteAddr().String(), nil
}

func (this *Server) Addr() string {
	return this.Listener.Addr().String()
}
