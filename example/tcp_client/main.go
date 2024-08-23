package main

import (
	"github.com/injoyai/ios"
	"github.com/injoyai/ios/module/client"
	"github.com/injoyai/ios/module/client/dial"
	"github.com/injoyai/ios/module/tcp"
	"time"
)

func main() {
	dial.RedialTCP(":10086", func(c *client.Client) {
		c.GoTimerWriter(time.Second*3, func(w ios.MoreWriter) error {
			return w.WriteAny(time.Now().Format("2006-01-02 15:04:05"))
		})
	}).Run()

	client.Redial(tcp.NewDial(":10086"), func(c *client.Client) {
		c.GoTimerWriter(time.Second*3, func(w ios.MoreWriter) error {
			return w.WriteAny(time.Now().Format("2006-01-02 15:04:05"))
		})
	}).Run()

	client.Run(tcp.NewDial(":10086"), func(c *client.Client) {
		c.SetRedial()
		c.GoTimerWriter(time.Second*3, func(w ios.MoreWriter) error {
			return w.WriteAny(time.Now().Format("2006-01-02 15:04:05"))
		})
	})

}
