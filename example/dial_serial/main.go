package main

import (
	"bufio"
	"context"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/ios/client/dial"
	"github.com/injoyai/ios/module/serial"
	"github.com/injoyai/logs"
	"os"
	"time"
)

func main() {
	c, err := dial.Serial(&serial.Config{
		Address:  "COM2",
		BaudRate: 9600,
		DataBits: 8,
		StopBits: 1,
		Parity:   "N",
		Timeout:  time.Second * 10,
	}, func(c *client.Client) {
		c.Event.OnDealErr = func(c *client.Client, err error) error {
			if err != nil && err.Error() == "serial: timeout" {
				return nil
			}
			return err
		}
	})
	logs.PanicErr(err)

	go func() {
		buf := bufio.NewReader(os.Stdin)
		for {
			bs, _, _ := buf.ReadLine()
			bs = append(bs, '\r', '\n')
			c.Write(bs)
		}
	}()
	c.Run(context.Background())
}
