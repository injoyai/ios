package dial

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

var (
	WithMemory    = memory.NewDial
	WithMQTT      = mqtt.NewDial
	WithRabbitmq  = rabbitmq.Dial
	WithSerial    = serial.NewDial
	WithSSH       = ssh.NewDial
	WithTCP       = tcp.NewDial
	WithWebsocket = websocket.NewDial
)

func Dial(dial ios.DialFunc, op ...client.Option) (*client.Client, error) {
	return client.Dial(dial, op...)
}

func Run(dial ios.DialFunc, op ...client.Option) error {
	c, err := Dial(dial, op...)
	if err != nil {
		return err
	}
	return c.Run(context.Background())
}

func TCP(addr string, op ...client.Option) (*client.Client, error) {
	return client.Dial(tcp.NewDial(addr), op...)
}

func RunTCP(addr string, op ...client.Option) error {
	return Run(tcp.NewDial(addr), op...)
}

func SSH(cfg *ssh.Config, op ...client.Option) (*client.Client, error) {
	return client.Dial(ssh.NewDial(cfg), op...)
}

func RunSSH(cfg *ssh.Config, op ...client.Option) error {
	return Run(ssh.NewDial(cfg), op...)
}

func Websocket(addr string, op ...client.Option) (*client.Client, error) {
	return client.Dial(websocket.NewDial(addr), op...)
}

func RunWebsocket(addr string, op ...client.Option) error {
	return Run(websocket.NewDial(addr), op...)
}

func Serial(cfg *serial.Config, op ...client.Option) (*client.Client, error) {
	return client.Dial(serial.NewDial(cfg), op...)
}

func RunSerial(cfg *serial.Config, op ...client.Option) error {
	return Run(serial.NewDial(cfg), op...)
}

func MQTT(cfg *mqtt.Config, subscribe mqtt.Subscribe, publish mqtt.Publish, op ...client.Option) (*client.Client, error) {
	return client.Dial(mqtt.NewDial(cfg, subscribe, publish), op...)
}

func RunMQTT(cfg *mqtt.Config, subscribe mqtt.Subscribe, publish mqtt.Publish, op ...client.Option) error {
	return Run(mqtt.NewDial(cfg, subscribe, publish), op...)
}

func Rabbitmq(addr string, cfg *rabbitmq.Config, op ...client.Option) (*client.Client, error) {
	return client.Dial(rabbitmq.NewDial(addr, cfg), op...)
}

func RunRabbitmq(addr string, cfg *rabbitmq.Config, op ...client.Option) error {
	return Run(rabbitmq.NewDial(addr, cfg), op...)
}

func Memory(key string, op ...client.Option) (*client.Client, error) {
	return client.Dial(memory.NewDial(key), op...)
}

func RunMemory(key string, op ...client.Option) error {
	return Run(memory.NewDial(key), op...)
}
