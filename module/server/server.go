package server

import (
	"context"
	"github.com/injoyai/base/safe"
	"github.com/injoyai/ios"
	"github.com/injoyai/ios/module/client"
	"io"
	"net"
)

type Listener interface {
	io.Closer
	Accept() (ios.ReadWriteCloser, string, error)
	Addr() string
}

func New(network, address string) (*Server, error) {
	return NewWithContext(context.Background(), network, address)
}

func NewWithContext(ctx context.Context, network, address string) (*Server, error) {
	listen, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	s := &Server{
		Closer:   safe.NewCloser(),
		Runner:   safe.NewRunnerWithContext(ctx, nil),
		ctx:      ctx,
		listener: listen,
	}
	s.Runner.SetFunc(s.run)
	return s, nil
}

type Server struct {
	*safe.Closer
	*safe.Runner

	ctx      context.Context //ctx
	listener net.Listener    //listener
}

func (this *Server) run(ctx context.Context) error {
	for {
		c, err := this.listener.Accept()
		if err != nil {
			return err
		}
		go func() {
			cli := client.NewWithContext(ctx)
			cli.SetReadWriteCloser(c.RemoteAddr().String(), c)
			cli.Run()
		}()
	}
}
