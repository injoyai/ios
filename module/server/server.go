package server

import (
	"context"
	"github.com/injoyai/base/safe"
	"github.com/injoyai/ios"
	"github.com/injoyai/ios/module/client"
	"io"
)

type Listener interface {
	io.Closer
	Accept() (ios.ReadWriteCloser, string, error)
	Addr() string
}

type Server struct {
	*safe.Closer
	*safe.Runner

	ctx      context.Context //ctx
	listener Listener        //listener
}

func (this *Server) run(ctx context.Context) error {
	for {
		c, k, err := this.listener.Accept()
		if err != nil {
			return err
		}
		go func() {

			cli := client.New(nil)
			cli.Reader = c
			cli.MoreWriter = ios.NewMoreWriter(c)
			this.Runner = safe.NewRunnerWithContext(this.ctx, this.run)
			cli.SetKey(k)
			cli.Run()

		}()
	}
}
