package dial

import (
	"github.com/injoyai/ios"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/ios/module/hid"
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
	WithHID       = hid.NewDial
)

func Dial(dial ios.DialFunc, op ...client.Option) (*client.Client, error) {
	return client.Dial(dial, op...)
}

func Redial(dial ios.DialFunc, op ...client.Option) *client.Client {
	return client.Redial(dial, op...)
}

func Run(dial ios.DialFunc, op ...client.Option) error {
	return Redial(dial, op...).Run()
}

func TCP(addr string, op ...client.Option) (*client.Client, error) {
	return client.Dial(tcp.NewDial(addr), op...)
}

func RedialTCP(addr string, op ...client.Option) *client.Client {
	return client.Redial(tcp.NewDial(addr), op...)
}

func RunTCP(addr string, op ...client.Option) error {
	return RedialTCP(addr, op...).Run()
}

func SSH(cfg *ssh.Config, op ...client.Option) (*client.Client, error) {
	return client.Dial(ssh.NewDial(cfg), op...)
}

func RedialSSH(cfg *ssh.Config, op ...client.Option) *client.Client {
	return client.Redial(ssh.NewDial(cfg), op...)
}

func RunSSH(cfg *ssh.Config, op ...client.Option) error {
	return RedialSSH(cfg, op...).Run()
}

func Websocket(addr string, op ...client.Option) (*client.Client, error) {
	return client.Dial(websocket.NewDial(addr), op...)
}

func RedialWebsocket(addr string, op ...client.Option) *client.Client {
	return client.Redial(websocket.NewDial(addr), op...)
}

func RunWebsocket(addr string, op ...client.Option) error {
	return RedialWebsocket(addr, op...).Run()
}

func Serial(cfg *serial.Config, op ...client.Option) (*client.Client, error) {
	return client.Dial(serial.NewDial(cfg), op...)
}

func RedialSerial(cfg *serial.Config, op ...client.Option) *client.Client {
	return client.Redial(serial.NewDial(cfg), op...)
}

func RunSerial(cfg *serial.Config, op ...client.Option) error {
	return RedialSerial(cfg, op...).Run()
}

func MQTT(cfg *mqtt.Config, subscribe mqtt.Subscribe, publish mqtt.Publish, op ...client.Option) (*client.Client, error) {
	return client.Dial(mqtt.NewDial(cfg, subscribe, publish), op...)
}

func RedialMQTT(cfg *mqtt.Config, subscribe mqtt.Subscribe, publish mqtt.Publish, op ...client.Option) *client.Client {
	return client.Redial(mqtt.NewDial(cfg, subscribe, publish), op...)
}

func RunMQTT(cfg *mqtt.Config, subscribe mqtt.Subscribe, publish mqtt.Publish, op ...client.Option) error {
	return RedialMQTT(cfg, subscribe, publish, op...).Run()
}

func Rabbitmq(addr string, cfg *rabbitmq.Config, op ...client.Option) (*client.Client, error) {
	return client.Dial(rabbitmq.NewDial(addr, cfg), op...)
}

func RedialRabbitmq(addr string, cfg *rabbitmq.Config, op ...client.Option) *client.Client {
	return client.Redial(rabbitmq.NewDial(addr, cfg), op...)
}

func RunRabbitmq(addr string, cfg *rabbitmq.Config, op ...client.Option) error {
	return RedialRabbitmq(addr, cfg, op...).Run()
}

func Memory(key string, op ...client.Option) (*client.Client, error) {
	return client.Dial(memory.NewDial(key), op...)
}

func RedialMemory(key string, op ...client.Option) *client.Client {
	return client.Redial(memory.NewDial(key), op...)
}

func RunMemory(key string, op ...client.Option) error {
	return RedialMemory(key, op...).Run()
}

func HID(vendorID, productID uint16, op ...client.Option) (*client.Client, error) {
	return client.Dial(hid.NewDial(vendorID, productID), op...)
}

func RedialHID(vendorID, productID uint16, op ...client.Option) *client.Client {
	return client.Redial(hid.NewDial(vendorID, productID), op...)
}

func RunHID(vendorID, productID uint16, op ...client.Option) error {
	return RedialHID(vendorID, productID, op...).Run()
}
