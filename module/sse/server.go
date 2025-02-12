package sse

import (
	"errors"
	"fmt"
	"github.com/injoyai/base/safe"
	"github.com/injoyai/ios"
	"net"
	"net/http"
)

func NewListen(port int) func() (ios.Listener, error) {
	return func() (ios.Listener, error) {
		l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			return nil, err
		}
		s := &Server{
			addr:      l.Addr().String(),
			closeFunc: l.Close,
			ch:        make(chan *Args),
		}
		return s, http.Serve(l, s)
	}

}

// NewHandlerListen is a good idea?
func NewHandlerListen(f func(h http.Handler)) func() (ios.Listener, error) {
	return func() (ios.Listener, error) {
		s := &Server{ch: make(chan *Args)}
		f(s)
		return s, nil
	}
}

type Server struct {
	addr      string
	closeFunc func() error
	ch        chan *Args
}

func (this *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	closer := safe.NewCloser().SetCloseFunc(func(err error) error {
		return r.Body.Close()
	})
	this.ch <- &Args{
		Request: r,
		Writer:  w,
		Closer:  closer,
	}
	<-closer.Done()
}

func (this *Server) Close() error {
	if this.closeFunc != nil {
		return this.closeFunc()
	}
	close(this.ch)
	return nil
}

func (this *Server) Accept() (ios.ReadWriteCloser, string, error) {
	i, ok := <-this.ch
	if !ok {
		return nil, "", errors.New("closed")
	}
	return i, fmt.Sprintf("%p", i), nil
}

func (this *Server) Addr() string {
	return this.addr
}

type Args struct {
	Request *http.Request
	Writer  http.ResponseWriter
	*safe.Closer
}

func (this *Args) Read(p []byte) (int, error) {
	return this.Request.Body.Read(p)
}

func (this *Args) Write(p []byte) (int, error) {
	n, err := this.Writer.Write(p)
	if err != nil {
		this.Close()
		return 0, err
	}
	return n, nil
}
