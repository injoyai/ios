package main

import (
	"context"
	"github.com/injoyai/ios"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/ios/client/dial"
	"github.com/injoyai/ios/client/frame"
	"github.com/injoyai/ios/server"
	"github.com/injoyai/ios/server/listen"
	"github.com/injoyai/logs"
	"io"
	"time"
)

func main() {

	go listen.RunTCP(8080, func(s *server.Server) {
		s.SetClientOption(func(c *client.Client) {
			c.Logger.Debug(false)
			c.Event.WithFrame(frame.Default)
			c.Event.OnDealMessage = func(c *client.Client, msg ios.Acker) {
				c.Write(msg.Payload())
			}
		})
	})

	dial.RedialTCP("127.0.0.1:8080", func(c *client.Client) {
		logs.Debug(c.Event.OnReadFrom)
		c.Event.OnWriteWith = frame.Default.WriteWith
		c.Event.OnReadFrom = func(r io.Reader) ([]byte, error) {
			logs.Debug("ReadFrom")
			return frame.Default.ReadFrom(r)
		}
		c.GoTimerWriter(time.Second, func(w ios.MoreWriter) error {
			return w.WriteAny(time.Now())
		})
	}).Run(context.Background())

}
