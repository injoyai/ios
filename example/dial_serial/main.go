package main

import (
	"bufio"
	"context"
	"os"
	"time"

	"github.com/injoyai/ios/v2/client"
	"github.com/injoyai/ios/v2/client/dial"
	"github.com/injoyai/ios/v2/module/serial"
	"github.com/injoyai/logs"
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
		c.OnDealErr(func(c *client.Client, err error) error {
			if err != nil && err.Error() == "serial: timeout" {
				return nil
			}
			return err
		})
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
