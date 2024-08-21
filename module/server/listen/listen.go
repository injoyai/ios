package listen

import (
	"github.com/injoyai/ios/module/memory"
	"github.com/injoyai/ios/module/server"
	"github.com/injoyai/ios/module/tcp"
	"github.com/injoyai/ios/module/websocket"
)

func TCP(port int, op ...server.Option) (*server.Server, error) {
	return server.New(tcp.NewListen(port), op...)
}

func TCPRun(port int, op ...server.Option) error {
	return server.Run(tcp.NewListen(port), op...)
}

func Memory(key string, op ...server.Option) (*server.Server, error) {
	return server.New(memory.NewListen(key), op...)
}

func MemoryRun(key string, op ...server.Option) error {
	return server.Run(memory.NewListen(key), op...)
}

func Websocket(port int, op ...server.Option) (*server.Server, error) {
	return server.New(websocket.NewListen(port), op...)
}

func WebsocketRun(port int, op ...server.Option) error {
	return server.Run(websocket.NewListen(port), op...)
}
