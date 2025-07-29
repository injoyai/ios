package main

import (
	"context"
	"github.com/injoyai/ios"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/ios/client/frame"
	"github.com/injoyai/ios/client/redial"
	"github.com/injoyai/ios/server"
	"github.com/injoyai/ios/server/listen"
	"time"
)

func main() {

	go listen.RunTCP(10089, func(s *server.Server) {
		s.SetClientOption(func(c *client.Client) {
			c.Event.WithFrame(frame.Default)
		})
	})

	redial.TCP(":10089", func(c *client.Client) {
		c.Event.WithFrame(frame.Default)
		c.GoTimerWriter(time.Second*5, func(w ios.MoreWriter) error {
			return w.WriteAny(time.Now().String())
		})
	}).Run(context.Background())

}
