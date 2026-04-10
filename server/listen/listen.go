package listen

import (
	"github.com/injoyai/ios/v2"
	"github.com/injoyai/ios/v2/module/common"
	"github.com/injoyai/ios/v2/module/memory"
	"github.com/injoyai/ios/v2/module/tcp"
	"github.com/injoyai/ios/v2/module/udp"
	"github.com/injoyai/ios/v2/module/unix"
	"github.com/injoyai/ios/v2/module/websocket"
	"github.com/injoyai/ios/v2/server"
)

func Run(listen ios.ListenFunc, op ...server.Option) error {
	return server.Run(listen, op...)
}

func TCP[T common.Address](addr T, op ...server.Option) (*server.Server, error) {
	return server.New(tcp.NewListen(addr), op...)
}

func RunTCP[T common.Address](addr T, op ...server.Option) error {
	return server.Run(tcp.NewListen(addr), op...)
}

func UDP[T common.Address](addr T, op ...server.Option) (*server.Server, error) {
	return server.New(udp.NewListen(addr), op...)
}

func RunUDP[T common.Address](addr T, op ...server.Option) error {
	return server.Run(udp.NewListen(addr), op...)
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

func Websocket[T common.Address](addr T, op ...server.Option) (*server.Server, error) {
	return server.New(websocket.NewListen(addr), op...)
}

func RunWebsocket[T common.Address](addr T, op ...server.Option) error {
	return server.Run(websocket.NewListen(addr), op...)
}
