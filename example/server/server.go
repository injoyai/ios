package main

import (
	"github.com/injoyai/ios/v2/module/tcp"
	"github.com/injoyai/ios/v2/server"
	"github.com/injoyai/logs"
)

func main() {
	err := server.Run(tcp.NewListen(10086))
	logs.PanicErr(err)
}
