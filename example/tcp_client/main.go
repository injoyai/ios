package main

import (
	"context"
	"github.com/injoyai/ios"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/ios/client/dial"
	"github.com/injoyai/ios/module/tcp"
	"time"
)

func main() {
	dial.RedialTCP(":10086", func(c *client.Client) {
		c.GoTimerWriter(time.Second*3, func(w ios.MoreWriter) error {
			return w.WriteAny(time.Now().Format("2006-01-02 15:04:05"))
		})
	}).Run(context.Background())

	client.Redial(tcp.NewDial(":10086"), func(c *client.Client) {
		c.GoTimerWriter(time.Second*3, func(w ios.MoreWriter) error {
			return w.WriteAny(time.Now().Format("2006-01-02 15:04:05"))
		})
	}).Run(context.Background())

	client.Run(tcp.NewDial(":10086"), func(c *client.Client) {
		c.SetRedial()
		c.GoTimerWriter(time.Second*3, func(w ios.MoreWriter) error {
			return w.WriteAny(time.Now().Format("2006-01-02 15:04:05"))
		})
	})

}
