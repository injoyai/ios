package server

import (
	"github.com/injoyai/ios"
)

type Listener interface {
	Accept() (ios.ReadWriteCloser, string, error)
}

type Server struct {
}
