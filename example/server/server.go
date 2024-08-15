package main

import (
	"github.com/injoyai/ios/module/server"
	"github.com/injoyai/logs"
)

func main() {
	s, err := server.New("tcp", ":10086")
	logs.PanicErr(err)
	s.Run()
}
