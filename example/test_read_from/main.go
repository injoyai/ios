package main

import (
	"context"
	"io"
	"time"

	"github.com/injoyai/ios"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/ios/client/frame"
	"github.com/injoyai/ios/client/redial"
	"github.com/injoyai/ios/server"
	"github.com/injoyai/ios/server/listen"
	"github.com/injoyai/logs"
)

func main() {

	go listen.RunTCP(8080, func(s *server.Server) {
		s.SetClientOption(func(c *client.Client) {
			c.Logger.Debug(false)
			c.WithFrame(frame.Default)
			c.OnDealMessage(func(c *client.Client, msg ios.Acker) {
				c.Write(msg.Bytes())
			})
		})
	})

	redial.TCP("127.0.0.1:8080", func(c *client.Client) {
		c.OnWriteWith(frame.Default.WriteWith)
		c.OnReadFrom(func(r io.Reader) ([]byte, error) {
			logs.Debug("ReadFrom")
			return frame.Default.ReadFrom(r)
		})
		c.GoTimerWriter(time.Second, func(w ios.MoreWriter) error {
			return w.WriteAny(time.Now())
		})
	}).Run(context.Background())

}
