package mqtt

import (
	"context"
	"errors"
	"fmt"
	"github.com/DrmagicE/gmqtt"
	"github.com/DrmagicE/gmqtt/pkg/packets"
	"github.com/DrmagicE/gmqtt/server"
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

		ch := make(chan *Conn)
		srv := server.New(server.WithTCPListener(l))
		if err := srv.Init(server.WithHook(server.Hooks{
			OnConnected: func(ctx context.Context, client server.Client) {
				//订阅clientID,conn
				srv.SubscriptionService().Subscribe(client.ClientOptions().ClientID, &gmqtt.Subscription{
					TopicFilter: client.ClientOptions().ClientID,
					QoS:         packets.Qos0,
				})
				ch <- &Conn{
					ClientID:  client.ClientOptions().ClientID,
					Client:    client,
					Publisher: srv.Publisher(),
				}
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

		go srv.Run()

		s := &Server{
			addr: l.Addr().String(),
			ch:   ch,
			stop: srv.Stop,
		}

		return s, nil
	}
}

type Conn struct {
	ClientID string
	server.Client
	server.Publisher
}

func (c *Conn) Write(p []byte) (n int, err error) {
	c.Publisher.Publish(&gmqtt.Message{
		Topic:   c.ClientID,
		Payload: p,
	})
	return len(p), nil
}

func (c *Conn) Close() error {
	c.Client.Close()
	return nil
}

type Server struct {
	addr string
	ch   chan *Conn
	stop func(ctx context.Context) error
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
