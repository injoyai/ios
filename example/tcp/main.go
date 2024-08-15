package main

import (
	"github.com/injoyai/ios"
	"github.com/injoyai/ios/module/client"
	"github.com/injoyai/ios/module/tcp"
	"time"
)

func main() {

	if false {

		c := client.MustDial(tcp.NewDial(":10086"), func(c *client.Client) {
			go c.TimerWriter(time.Second*3, func(w ios.MoreWriter) error {
				return w.WriteAny(time.Now().Format("2006-01-02 15:04:05"))
			})
		})

		c.Run()
	}

	{

		c := client.New(tcp.NewDial(":10086"))
		c.Event.OnReconnect = client.WithReconnectInterval(time.Second * 3)
		c.Dial(true, func(c *client.Client) {
			c.SetKey(":10087")
			c.SetRedial()
			//c.SetReadTimeout(time.Second * 10)
			go c.TimerWriter(time.Second*3, func(w ios.MoreWriter) error {
				return w.WriteAny(time.Now().Format("2006-01-02 15:04:05"))
			})
		})
		c.Run()
	}

}
