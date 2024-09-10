package main

import (
	"github.com/injoyai/ios/client"
	"github.com/injoyai/ios/server"
	"github.com/injoyai/ios/server/listen"
	"github.com/injoyai/logs"
	"time"
)

func main() {
	logs.Err(listen.RunMQTT(11883, func(s *server.Server) {
		go s.Timer(time.Second*5, func(s *server.Server) {
			s.RangeClient(func(c *client.Client) bool {
				c.WriteAny(time.Now().String())
				return true
			})
		})
	}))
}
