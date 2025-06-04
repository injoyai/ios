package main

import (
	"github.com/injoyai/ios/client"
	"github.com/injoyai/ios/client/dial"
)

func main() {
	dial.RedialHID(0x045e, 0x028e, func(c *client.Client) {
		c.Logger.WithHEX()
	}).Run()
}
