package main

import (
	"time"

	"github.com/injoyai/ios/v2/server"
	"github.com/injoyai/ios/v2/server/listen"
)

func main() {
	listen.RunTCP(20088, func(s *server.Server) {
		s.SetTimeout(10*time.Second, 3*time.Second)
	})
}
