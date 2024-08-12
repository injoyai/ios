package mqtt

import (
	"context"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/injoyai/base/safe"
	"github.com/injoyai/conv"
	"github.com/injoyai/ios"
	"io"
	"strings"
	"time"
)

var _ ios.AReadWriteCloser = &Client{}

type (
	ClientOptions = mqtt.ClientOptions
	MQTTClient    = mqtt.Client
)

func NewDial(cfg *ClientOptions, subscribe Subscribe, publish Publish) ios.DialFunc {
	return func(ctx context.Context) (ios.ReadWriteCloser, string, error) {
		c, err := DialClient(cfg, subscribe, publish)
		key := cfg.ClientID
		if len(cfg.Servers) > 0 && cfg.Servers[0] != nil {
			key = cfg.Servers[0].Host
		}
		return c, key, err
	}
}

func DialClient(cfg *ClientOptions, subscribe Subscribe, publish Publish) (*Client, error) {
	c := mqtt.NewClient(cfg)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return Dial(c, subscribe, publish)
}

func Dial(c MQTTClient, subscribe Subscribe, publish Publish) (*Client, error) {
	cli := &Client{
		Client:    c,
		sub:       make(chan *Message),
		subscribe: subscribe,
		publish:   publish,
		Closer:    safe.NewCloser(),
	}

	cli.Closer.SetCloseFunc(func(err error) error {
		token := c.Unsubscribe(subscribe.Topic)
		token.Wait()
		if err := token.Error(); err != nil {
			return err
		}
		close(cli.sub)
		return nil
	})

	c.Subscribe(subscribe.Topic, subscribe.Qos, func(client mqtt.Client, msg mqtt.Message) {
		cli.sub <- &Message{msg}
	})

	return cli, nil
}

type Client struct {
	mqtt.Client
	sub       chan *Message
	subscribe Subscribe
	publish   Publish
	*safe.Closer
}

func (this *Client) ReadAck() (ios.Acker, error) {
	if this.Closed() {
		return nil, this.Closer.Err()
	}
	m, ok := <-this.sub
	if !ok {
		return nil, io.EOF
	}
	return m, nil
}

func (this *Client) Write(p []byte) (n int, err error) {
	if this.Closed() {
		return 0, this.Closer.Err()
	}
	token := this.Client.Publish(this.publish.Topic, this.publish.Qos, this.publish.Retained, p)
	token.Wait()
	return len(p), token.Error()
}

func (this *Client) Close() error {
	return this.Closer.CloseWithErr(io.EOF)
}

type Message struct {
	mqtt.Message
}

func (this *Message) Ack() error {
	this.Message.Ack()
	return nil
}

type Publish struct {
	Topic    string
	Qos      uint8
	Retained bool
}

type Subscribe struct {
	Topic string
	Qos   uint8
}

type BaseConfig struct {
	BrokerURL      string        //必选,不要忘记 tcp://
	ClientID       string        //必选,服务器Topic地址
	Username       string        //用户名
	Password       string        //密码
	ConnectTimeout time.Duration //连接超时时间,
	KeepAlive      time.Duration //心跳时间,0是不启用该机制
}

func WithBase(cfg *BaseConfig) *mqtt.ClientOptions {
	if !strings.HasPrefix(cfg.BrokerURL, "tcp://") {
		cfg.BrokerURL = "tcp://" + cfg.BrokerURL
	}
	if len(cfg.ClientID) == 0 {
		cfg.ClientID = conv.String(time.Now().UnixNano())
	}

	return mqtt.NewClientOptions().
		AddBroker(cfg.BrokerURL).
		SetClientID(cfg.ClientID).
		SetUsername(cfg.Username).
		SetPassword(cfg.Password).
		SetConnectTimeout(cfg.ConnectTimeout).
		SetKeepAlive(cfg.KeepAlive).
		SetAutoReconnect(false). //自动重连
		SetCleanSession(false)   //重连后恢复session
}
