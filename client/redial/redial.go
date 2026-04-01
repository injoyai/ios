package redial

import (
	"github.com/injoyai/ios/v2"
	"github.com/injoyai/ios/v2/client"
	"github.com/injoyai/ios/v2/module/memory"
	"github.com/injoyai/ios/v2/module/tcp"
	"github.com/injoyai/ios/v2/module/unix"
	"github.com/injoyai/ios/v2/module/websocket"
)

func With(dial ios.DialFunc, op ...client.Option) *client.Client {
	return client.Redial(dial, op...)
}

func Run(dial ios.DialFunc, op ...client.Option) error {
	return With(dial, op...).Run()
}

func TCP(addr string, op ...client.Option) *client.Client {
	return With(tcp.NewDial(addr), op...)
}

func RunTCP(addr string, op ...client.Option) error {
	return TCP(addr, op...).Run()
}

func Unix(addr string, op ...client.Option) *client.Client {
	return With(unix.NewDial(addr), op...)
}

func RunUnix(addr string, op ...client.Option) error {
	return Unix(addr, op...).Run()
}

func Websocket(addr string, op ...client.Option) *client.Client {
	return With(websocket.NewDial(addr), func(c *client.Client) {
		c.OnWrite(client.NewWriteSafe())
		c.SetOption(op...)
	})
}

func RunWebsocket(addr string, op ...client.Option) error {
	return Websocket(addr, op...).Run()
}

func Memory(key string, op ...client.Option) *client.Client {
	return With(memory.NewDial(key), op...)
}

func RunMemory(key string, op ...client.Option) error {
	return Memory(key, op...).Run()
}
