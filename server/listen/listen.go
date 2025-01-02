package listen

import (
	"context"
	"github.com/injoyai/ios/module/memory"
	"github.com/injoyai/ios/module/mqtt"
	"github.com/injoyai/ios/module/tcp"
	"github.com/injoyai/ios/module/websocket"
	"github.com/injoyai/ios/server"
)

func TCP(port int, op ...server.Option) (*server.Server, error) {
	return server.New(tcp.NewListen(port), op...)
}

func TCPContext(ctx context.Context, port int, op ...server.Option) (*server.Server, error) {
	return server.New(tcp.NewListenContext(ctx, port), op...)
}

func RunTCP(port int, op ...server.Option) error {
	return server.Run(tcp.NewListen(port), op...)
}

func RunTCPContext(ctx context.Context, port int, op ...server.Option) error {
	return server.Run(tcp.NewListenContext(ctx, port), op...)
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

func MQTT(port int, op ...server.Option) (*server.Server, error) {
	return server.New(mqtt.NewListen(port), op...)
}

func RunMQTT(port int, op ...server.Option) error {
	return server.Run(mqtt.NewListen(port), op...)
}
