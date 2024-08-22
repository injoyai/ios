package mqtt

import (
	"context"
	"errors"
	"fmt"
	"github.com/DrmagicE/gmqtt"
	_ "github.com/DrmagicE/gmqtt/persistence"
	"github.com/DrmagicE/gmqtt/pkg/packets"
	"github.com/DrmagicE/gmqtt/server"
	_ "github.com/DrmagicE/gmqtt/topicalias/fifo"
	"github.com/injoyai/base/maps"
	"github.com/injoyai/base/safe"
	"github.com/injoyai/ios"
	"net"
)

func NewListen(port int) ios.ListenFunc {
	return func() (ios.Listener, error) {
		l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			return nil, err
		}
		return NewNetListen(l)()
	}
}

func NewNetListen(l net.Listener) ios.ListenFunc {
	return func() (ios.Listener, error) {

		s := &Server{
			addr:   l.Addr().String(),
			ch:     make(chan *Conn),
			closer: safe.NewCloser(),
			m:      maps.NewSafe(),
		}

		srv := server.New(server.WithTCPListener(l))
		if err := srv.Init(server.WithHook(server.Hooks{
			OnConnected: func(ctx context.Context, client server.Client) {
				//订阅clientID,conn
				clientID := client.ClientOptions().ClientID
				srv.SubscriptionService().Subscribe(clientID, &gmqtt.Subscription{
					TopicFilter: clientID,
					QoS:         packets.Qos0,
				})
				conn := &Conn{
					ClientID:  clientID,
					Client:    client,
					Publisher: srv.Publisher(),
					Closer: safe.NewCloser().SetCloseFunc(func(err error) error {
						client.Close()
						return nil
					}),
					ch: make(chan []byte),
				}
				s.ch <- conn
				s.m.Set(clientID, conn)
			},
			OnClosed: func(ctx context.Context, client server.Client, err error) {
				//取消订阅
				clientID := client.ClientOptions().ClientID
				srv.SubscriptionService().UnsubscribeAll(clientID)
				if conn, _ := s.m.GetAndDel(clientID); conn != nil {
					conn.(*Conn).CloseWithErr(err)
				}
			},
			OnMsgArrived: func(ctx context.Context, client server.Client, msg *server.MsgArrivedRequest) error {
				clientID := client.ClientOptions().ClientID
				if conn := s.m.MustGet(clientID); conn != nil {
					conn.(*Conn).ch <- msg.Message.Payload
				}
				return nil
			},
			OnSubscribe: func(ctx context.Context, client server.Client, req *server.SubscribeRequest) error {
				if req == nil || req.Subscribe == nil {
					return nil
				}
				for _, topic := range req.Subscribe.Topics {
					srv.SubscriptionService().Subscribe(client.ClientOptions().ClientID, &gmqtt.Subscription{
						TopicFilter: topic.Name,
						QoS:         topic.Qos,
					})
				}
				return nil
			},
		})); err != nil {
			return nil, err
		}

		s.stop = srv.Stop

		go func() {
			s.closer.CloseWithErr(srv.Run())
		}()

		return s, nil
	}
}

type Conn struct {
	ClientID  string
	Client    server.Client
	Publisher server.Publisher
	*safe.Closer
	ch chan []byte
}

func (c *Conn) ReadMessage() ([]byte, error) {
	select {
	case <-c.Closer.Done():
		return nil, c.Closer.Err()
	case bs, ok := <-c.ch:
		if !ok {
			return nil, errors.New("conn closed")
		}
		return bs, nil
	}
}

func (c *Conn) Write(p []byte) (n int, err error) {
	if c.Closer.Closed() {
		return 0, c.Closer.Err()
	}
	c.Publisher.Publish(&gmqtt.Message{
		Topic:   c.ClientID,
		Payload: p,
	})
	return len(p), nil
}

type Server struct {
	addr   string
	ch     chan *Conn
	stop   func(ctx context.Context) error
	closer *safe.Closer
	m      *maps.Safe
}

func (this *Server) Close() error {
	return this.stop(context.Background())
}

func (this *Server) Accept() (ios.ReadWriteCloser, string, error) {
	conn, ok := <-this.ch
	if !ok {
		return nil, "", errors.New("listen closed")
	}
	return conn, conn.Client.ClientOptions().ClientID, nil
}

func (this *Server) Addr() string {
	return this.addr
}
