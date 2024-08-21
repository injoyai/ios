package websocket

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/injoyai/base/safe"
	"github.com/injoyai/ios"
	"net"
	"net/http"
)

func NewListen(port int) func() (ios.Listener, error) {
	return func() (ios.Listener, error) {

		addr := fmt.Sprintf(":%d", port)
		l, err := net.Listen("tcp", addr)
		if err != nil {
			return nil, err
		}

		ch := make(chan *Conn)
		s := &Server{
			addr: addr,
			ch:   ch,
			srv: &http.Server{
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					up := websocket.Upgrader{
						ReadBufferSize:  1024,
						WriteBufferSize: 1024,
					}
					c, err := up.Upgrade(w, r, r.Header)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						return
					}

					conn := &Conn{
						Conn:   c,
						closer: safe.NewCloser(),
					}

					ch <- conn

					//这里加个异步退出
					<-conn.closer.Done()

					//关闭连接
					c.Close()

				})},
		}

		go s.srv.Serve(l)

		return s, nil
	}
}

type Conn struct {
	*websocket.Conn
	closer *safe.Closer
}

func (c *Conn) ReadMessage() ([]byte, error) {
	_, bs, err := c.Conn.ReadMessage()
	return bs, err
}

func (c *Conn) Write(p []byte) (int, error) {
	err := c.WriteMessage(websocket.BinaryMessage, p)
	return len(p), err
}

func (c *Conn) Close() error {
	return c.closer.Close()
}

type Server struct {
	addr string
	ch   chan *Conn
	srv  *http.Server
}

func (this *Server) Close() error {
	return this.srv.Close()
}

func (this *Server) Accept() (ios.ReadWriteCloser, string, error) {
	conn, ok := <-this.ch
	if !ok {
		return nil, "", errors.New("listen closed")
	}
	return conn, conn.RemoteAddr().String(), nil
}

func (this *Server) Addr() string {
	return this.addr
}
