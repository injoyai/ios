package main

import (
	"github.com/injoyai/ios"
	"github.com/injoyai/ios/module/client"
	"github.com/injoyai/ios/module/tcp"
	"time"
)

func main() {

	c := client.MustDial(tcp.NewDial(":10086"), func(c *client.Client) {
		c.SetRedial()
		go c.TimerWriter(time.Second*3, func(w ios.MoreWriter) error {
			return w.WriteAny(time.Now().Format("2006-01-02 15:04:05"))
		})
	})

	c.Run()
}
