package main

import (
	"github.com/injoyai/ios/module/tcp"
	"github.com/injoyai/ios/server"
	"github.com/injoyai/logs"
)

func main() {
	err := server.Run(tcp.NewListen(10086))
	logs.PanicErr(err)
}
