package dial

import (
	"github.com/injoyai/ios/v2"
	"github.com/injoyai/ios/v2/client"
	"github.com/injoyai/ios/v2/module/memory"
	"github.com/injoyai/ios/v2/module/tcp"
	"github.com/injoyai/ios/v2/module/unix"
	"github.com/injoyai/ios/v2/module/websocket"
)

func With(dial ios.DialFunc, op ...client.Option) (*client.Client, error) {
	return client.Dial(dial, op...)
}

func Run(dial ios.DialFunc, op ...client.Option) error {
	c, err := With(dial, op...)
	if err != nil {
		return err
	}
	return c.Run()
}

func TCP(addr string, op ...client.Option) (*client.Client, error) {
	return With(tcp.NewDial(addr), op...)
}

func RunTCP(addr string, op ...client.Option) error {
	return Run(tcp.NewDial(addr), op...)
}

func Unix(addr string, op ...client.Option) (*client.Client, error) {
	return With(unix.NewDial(addr), op...)
}

func RunUnix(addr string, op ...client.Option) error {
	return Run(unix.NewDial(addr), op...)
}

func Websocket(addr string, op ...client.Option) (*client.Client, error) {
	return With(websocket.NewDial(addr), func(c *client.Client) {
		c.OnWrite(client.NewWriteSafe())
		c.SetOption(op...)
	})
}

func RunWebsocket(addr string, op ...client.Option) error {
	return Run(websocket.NewDial(addr), func(c *client.Client) {
		c.OnWrite(client.NewWriteSafe())
		c.SetOption(op...)
	})
}

func Memory(key string, op ...client.Option) (*client.Client, error) {
	return With(memory.NewDial(key), op...)
}

func RunMemory(key string, op ...client.Option) error {
	return Run(memory.NewDial(key), op...)
}
