package main

import (
	"github.com/injoyai/ios"
	"github.com/injoyai/ios/module/client"
	"github.com/injoyai/ios/module/server"
	"github.com/injoyai/ios/module/server/listen"
	"github.com/injoyai/logs"
	"time"
)

func main() {
	logs.Err(listen.WebsocketRun(18080, func(s *server.Server) {
		s.SetClientOption(func(c *client.Client) {
			c.GoTimerWriter(time.Second*5, func(w ios.MoreWriter) error {
				return w.WriteAny(time.Now().String())
			})
		})
	}))
}
