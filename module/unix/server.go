package unix

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/injoyai/ios/v2"
)

func NewListen(filename string) ios.ListenFunc {
	return func() (ios.Listener, error) {
		os.Remove(filename)
		os.MkdirAll(filepath.Dir(filename), 0755)
		listener, err := net.Listen("unix", filename)
		if err != nil {
			return nil, err
		}
		return &Server{
			Listener: listener,
			filename: filename,
		}, nil
	}
}

type Server struct {
	net.Listener
	filename string
}

func (this *Server) Close() error {
	os.Remove(this.filename)
	return this.Listener.Close()
}

func (this *Server) Accept() (ios.ReadWriteCloser, string, error) {
	c, err := this.Listener.Accept()
	if err != nil {
		return nil, "", err
	}
	return c, fmt.Sprintf("%p", c), nil
}

func (this *Server) Addr() string {
	return this.Listener.Addr().String()
}
