package redial

import (
	"context"
	"github.com/injoyai/ios"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/ios/module/memory"
	"github.com/injoyai/ios/module/mqtt"
	"github.com/injoyai/ios/module/rabbitmq"
	"github.com/injoyai/ios/module/serial"
	"github.com/injoyai/ios/module/ssh"
	"github.com/injoyai/ios/module/tcp"
	"github.com/injoyai/ios/module/websocket"
)

func Redial(dial ios.DialFunc, op ...client.Option) *client.Client {
	return client.Redial(dial, op...)
}

func Run(dial ios.DialFunc, op ...client.Option) error {
	return Redial(dial, op...).Run(context.Background())
}

func TCP(addr string, op ...client.Option) *client.Client {
	return client.Redial(tcp.NewDial(addr), op...)
}

func RunTCP(addr string, op ...client.Option) error {
	return TCP(addr, op...).Run(context.Background())
}

func SSH(cfg *ssh.Config, op ...client.Option) *client.Client {
	return client.Redial(ssh.NewDial(cfg), op...)
}

func RunSSH(cfg *ssh.Config, op ...client.Option) error {
	return SSH(cfg, op...).Run(context.Background())
}

func Websocket(addr string, op ...client.Option) *client.Client {
	return client.Redial(websocket.NewDial(addr), func(c *client.Client) {
		c.OnWrite = client.NewWriteSafe()
		c.SetOption(op...)
	})
}

func RunWebsocket(addr string, op ...client.Option) error {
	return Websocket(addr, op...).Run(context.Background())
}

func Serial(cfg *serial.Config, op ...client.Option) *client.Client {
	return client.Redial(serial.NewDial(cfg), op...)
}

func RunSerial(cfg *serial.Config, op ...client.Option) error {
	return Serial(cfg, op...).Run(context.Background())
}

func MQTT(cfg *mqtt.Config, subscribe mqtt.Subscribe, publish mqtt.Publish, op ...client.Option) *client.Client {
	return client.Redial(mqtt.NewDial(cfg, subscribe, publish), op...)
}

func RunMQTT(cfg *mqtt.Config, subscribe mqtt.Subscribe, publish mqtt.Publish, op ...client.Option) error {
	return MQTT(cfg, subscribe, publish, op...).Run(context.Background())
}

func Rabbitmq(addr string, cfg *rabbitmq.Config, op ...client.Option) *client.Client {
	return client.Redial(rabbitmq.NewDial(addr, cfg), op...)
}

func RunRabbitmq(addr string, cfg *rabbitmq.Config, op ...client.Option) error {
	return Rabbitmq(addr, cfg, op...).Run(context.Background())
}

func Memory(key string, op ...client.Option) *client.Client {
	return client.Redial(memory.NewDial(key), op...)
}

func RunMemory(key string, op ...client.Option) error {
	return Memory(key, op...).Run(context.Background())
}
