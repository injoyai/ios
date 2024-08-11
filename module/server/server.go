package server

import (
	"context"
	"github.com/injoyai/base/safe"
	"github.com/injoyai/ios"
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
