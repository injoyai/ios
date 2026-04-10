package main

import (
	"time"

	"github.com/injoyai/ios/v2"
	"github.com/injoyai/ios/v2/client"
	"github.com/injoyai/ios/v2/client/redial"
	"github.com/injoyai/ios/v2/server/listen"
	"github.com/injoyai/logs"
)

func main() {

	go func() {
		err := listen.RunUDP(20087)
		logs.Err(err)
	}()

	<-time.After(time.Second)
	redial.RunUDP(":20087", func(c *client.Client) {
		c.OnConnected(func(c *client.Client) {
			c.GoTimerWriter(time.Second*3, func(w ios.MoreWriter) error {
				return w.WriteAny(time.Now().Format("2006-01-02 15:04:05"))
			})
		})
	})

}
