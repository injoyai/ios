package listen

import (
	"github.com/injoyai/ios/v2/module/memory"
	"github.com/injoyai/ios/v2/module/tcp"
	"github.com/injoyai/ios/v2/module/unix"
	"github.com/injoyai/ios/v2/module/websocket"
	"github.com/injoyai/ios/v2/server"
)

func TCP(port int, op ...server.Option) (*server.Server, error) {
	return server.New(tcp.NewListen(port), op...)
}

func RunTCP(port int, op ...server.Option) error {
	return server.Run(tcp.NewListen(port), op...)
}

func Unix(filename string, op ...server.Option) (*server.Server, error) {
	return server.New(unix.NewListen(filename), op...)
}

func RunUnix(filename string, op ...server.Option) error {
	return server.Run(unix.NewListen(filename), op...)
}

func Memory(key string, op ...server.Option) (*server.Server, error) {
	return server.New(memory.NewListen(key), op...)
}

func RunMemory(key string, op ...server.Option) error {
	return server.Run(memory.NewListen(key), op...)
}

func Websocket(port int, op ...server.Option) (*server.Server, error) {
	return server.New(websocket.NewListen(port), op...)
}

func RunWebsocket(port int, op ...server.Option) error {
	return server.Run(websocket.NewListen(port), op...)
}
