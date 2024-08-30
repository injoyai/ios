package dial

import (
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

func Redial(dial ios.DialFunc, op ...client.Option) *client.Client {
	return client.Redial(dial, op...)
}

func TCP(addr string, op ...client.Option) (*client.Client, error) {
	return client.Dial(tcp.NewDial(addr), op...)
}

func RedialTCP(addr string, op ...client.Option) *client.Client {
	return client.Redial(tcp.NewDial(addr), op...)
}

func SSH(cfg *ssh.Config, op ...client.Option) (*client.Client, error) {
	return client.Dial(ssh.NewDial(cfg), op...)
}

func RedialSSH(cfg *ssh.Config, op ...client.Option) *client.Client {
	return client.Redial(ssh.NewDial(cfg), op...)
}

func Websocket(addr string, op ...client.Option) (*client.Client, error) {
	return client.Dial(websocket.NewDial(addr), op...)
}

func RedialWebsocket(addr string, op ...client.Option) *client.Client {
	return client.Redial(websocket.NewDial(addr), op...)
}

func Serial(cfg *serial.Config, op ...client.Option) (*client.Client, error) {
	return client.Dial(serial.NewDial(cfg), op...)
}

func RedialSerial(cfg *serial.Config, op ...client.Option) *client.Client {
	return client.Redial(serial.NewDial(cfg), op...)
}

func MQTT(cfg *mqtt.Config, subscribe mqtt.Subscribe, publish mqtt.Publish, op ...client.Option) (*client.Client, error) {
	return client.Dial(mqtt.NewDial(cfg, subscribe, publish), op...)
}

func RedialMQTT(cfg *mqtt.Config, subscribe mqtt.Subscribe, publish mqtt.Publish, op ...client.Option) *client.Client {
	return client.Redial(mqtt.NewDial(cfg, subscribe, publish), op...)
}

func Rabbitmq(addr string, cfg *rabbitmq.Config, op ...client.Option) (*client.Client, error) {
	return client.Dial(rabbitmq.NewDial(addr, cfg), op...)
}

func RedialRabbitmq(addr string, cfg *rabbitmq.Config, op ...client.Option) *client.Client {
	return client.Redial(rabbitmq.NewDial(addr, cfg), op...)
}

func Memory(key string, op ...client.Option) (*client.Client, error) {
	return client.Dial(memory.NewDial(key), op...)
}

func RedialMemory(key string, op ...client.Option) *client.Client {
	return client.Redial(memory.NewDial(key), op...)
}
