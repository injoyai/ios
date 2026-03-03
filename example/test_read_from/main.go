package main

import (
	"bufio"
	"context"
	"time"

	"github.com/injoyai/ios/v2"
	"github.com/injoyai/ios/v2/client"
	"github.com/injoyai/ios/v2/client/frame"
	"github.com/injoyai/ios/v2/client/redial"
	"github.com/injoyai/ios/v2/server"
	"github.com/injoyai/ios/v2/server/listen"
	"github.com/injoyai/logs"
)

func main() {

	go listen.RunTCP(8080, func(s *server.Server) {
		s.OnConnected(func(c *client.Client) {
			c.Logger.Enable(false)
			c.WithFrame(frame.Default)
			c.OnDealMessage(func(c *client.Client, msg ios.Acker) {
				c.Write(msg.Bytes())
			})
		})
	})

	redial.TCP("127.0.0.1:8080", func(c *client.Client) {
		c.OnWriteWith(frame.Default.WriteWith)
		c.OnReadFrom(func(r *bufio.Reader) ([]byte, error) {
			logs.Debug("ReadFrom")
			return frame.Default.ReadFrom(r)
		})
		c.GoTimerWriter(time.Second, func(w ios.MoreWriter) error {
			return w.WriteAny(time.Now())
		})
	}).Run(context.Background())

}
