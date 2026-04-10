package udp

import (
	"fmt"
	"net"

	"github.com/injoyai/base/maps"
	"github.com/injoyai/base/safe"
	"github.com/injoyai/ios/v2"
)

var _ ios.Listener = (*Server)(nil)

func NewListen(port int) func() (ios.Listener, error) {
	return func() (ios.Listener, error) {
		addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
		if err != nil {
			return nil, err
		}
		conn, err := net.ListenUDP("udp", addr)
		if err != nil {
			return nil, err
		}
		s := &Server{
			conn:    conn,
			clients: maps.NewGeneric[string, *client](),
			accept:  make(chan *client, 10),
			Closer:  safe.NewCloser(),
		}
		s.Closer.SetCloseFunc(func(err error) error {
			s.conn.Close()
			close(s.accept)
			s.clients.Range(func(k string, v *client) bool {
				v.CloseWithErr(err)
				return true
			})
			return nil
		})
		go s.run()
		return s, nil
	}
}

type Server struct {
	conn    *net.UDPConn
	clients *maps.Generic[string, *client]
	accept  chan *client
	*safe.Closer
}

func (this *Server) Accept() (ios.ReadWriteCloser, string, error) {
	select {
	case <-this.Done():
		return nil, "", this.Err()
	case cli, ok := <-this.accept:
		if !ok {
			return nil, "", ios.ErrClosed
		}
		return cli, cli.key, nil
	}
}

func (this *Server) Addr() string {
	return this.conn.LocalAddr().String()
}

func (this *Server) run() error {
	buf := make([]byte, 65535)
	for {
		select {
		case <-this.Done():
			return this.Err()
		default:
		}

		n, addr, err := this.conn.ReadFromUDP(buf)
		if err != nil {
			return err
		}

		key := addr.String()
		conn, err := this.clients.GetOrSetByHandler(key, func() (*client, error) {
			c := newClient(key, this.conn, addr, func() { this.clients.Del(key) })
			select {
			case <-this.Done():
				return nil, this.Err()
			case this.accept <- c:
			}
			return c, nil
		})
		if err != nil {
			continue
		}

		bs := make([]byte, n)
		copy(bs, buf[:n])
		conn.addMessage(bs)

	}
}

var _ ios.MReadWriteCloser = (*client)(nil)

func newClient(key string, conn *net.UDPConn, addr *net.UDPAddr, onClose func()) *client {
	c := &client{
		key:        key,
		ch:         make(chan []byte, 10),
		conn:       conn,
		remoteAddr: addr,
		Closer:     safe.NewCloser(),
	}
	c.Closer.SetCloseFunc(func(err error) error {
		close(c.ch)
		onClose()
		return nil
	})
	return c
}

type client struct {
	key          string       //
	ch           chan []byte  //数据通道
	conn         *net.UDPConn //udp客户端
	remoteAddr   *net.UDPAddr //远程地址
	*safe.Closer              //
}

func (this *client) addMessage(bs []byte) {
	if this.Closed() {
		return
	}
	select {
	case this.ch <- bs:
	default:
	}
}

func (this *client) Write(p []byte) (int, error) {
	if this.Closed() {
		return 0, this.Err()
	}
	return this.conn.WriteToUDP(p, this.remoteAddr)
}

func (this *client) ReadMessage() ([]byte, error) {
	select {
	case <-this.Done():
		return nil, this.Err()
	case bs, ok := <-this.ch:
		if !ok {
			return nil, ios.ErrReadClosed
		}
		return bs, nil
	}
}
