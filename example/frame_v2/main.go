package main

import (
	"fmt"
	"github.com/injoyai/ios"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/ios/client/dial"
	"github.com/injoyai/ios/client/frame/v2"
	"github.com/injoyai/ios/server"
	"github.com/injoyai/ios/server/listen"
	"github.com/injoyai/logs"
	"time"
)

func main() {
	port := 10065

	go listen.RunTCP(port, func(s *server.Server) {
		s.SetClientOption(func(c *client.Client) {
			c.Event.WithFrame(frame.Default)
			//使用推荐结构
			c.OnDealMessage = frame.OnMessage(func(m *frame.Model) {
				logs.Info(m)
			})
			//使用默认(自定义)结构
			//c.OnDealMessage = func(c *client.Client, msg ios.Acker) {
			//	logs.Info(msg.Payload())
			//}
		})
	})

	dial.RedialTCP(fmt.Sprintf("127.0.0.1:%d", port), func(c *client.Client) {
		c.Event.WithFrame(frame.Default)
		c.Logger.WithHEX()
		c.GoTimerWriter(time.Second*5, func(w ios.MoreWriter) error {
			m := &frame.Model{
				Code:  frame.Succ,
				MsgID: 20,
				Type:  1,
				Data:  []byte{1, 2, 3},
			}
			return w.WriteAny(m.Bytes())
			//使用默认(自定义)结构
			return w.WriteAny("666")
		})
	}).Run()
}
