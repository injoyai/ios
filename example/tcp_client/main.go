package main

import (
	"context"
	"time"

	"github.com/injoyai/ios/v2"
	"github.com/injoyai/ios/v2/client"
	"github.com/injoyai/ios/v2/client/redial"
	"github.com/injoyai/ios/v2/module/tcp"
)

func main() {
	redial.TCP(":10086", func(c *client.Client) {
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
