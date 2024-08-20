package listen

import (
	"github.com/injoyai/ios/module/server"
	"github.com/injoyai/ios/module/tcp"
)

func RunTCP(port int, op ...server.Option) error {
	return server.Run(tcp.NewListen(port), op...)
}
