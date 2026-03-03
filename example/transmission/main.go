package main

import (
	"bufio"
	"os"
	"time"

	"github.com/injoyai/ios/v2"
	"github.com/injoyai/ios/v2/client"
	"github.com/injoyai/ios/v2/client/redial"
	"github.com/injoyai/ios/v2/module/common"
	"github.com/injoyai/ios/v2/server"
	"github.com/injoyai/ios/v2/server/listen"
	"github.com/injoyai/logs"
)

func main() {

	Test(0)
}

func Test(n int) {
	switch n {
	case 1:
		/*
			局域网测试结果:
			[调试]2023/08/03 15:03:32 main.go:52: [处理]传输耗时: 11.1MB/s
		*/
		logs.SetShowColor(false)
		var start time.Time  //当前时间
		length := 1000 << 20 //传输的数据大小
		totalDeal := 0
		listen.RunTCP(10086, func(s *server.Server) {
			s.Logger.SetLevel(common.LevelInfo)
			s.OnConnected(func(c *client.Client) {
				c.OnDealMessage(func(c *client.Client, msg ios.Acker) {
					defer msg.Ack()
					if start.IsZero() {
						start = time.Now()
					}
					totalDeal += len(msg.Bytes())
					if totalDeal >= length {
						logs.Debugf("[处理]传输耗时: %0.1fMB/s\n", float64(totalDeal/(1<<20))/time.Now().Sub(start).Seconds())
					}
				})
			})
		})
	case 0:
		/*
			测试结果:
			[调试]2023/08/03 15:03:30 main.go:62: [发送]传输耗时: 4507.1MB/s
			[调试]2023/08/03 15:03:32 main.go:25: [读取]传输耗时: 490.8MB
			[调试]2023/08/03 15:03:32 main.go:52: [处理]传输耗时: 490.7MB/s
		*/
		start := time.Now()  //当前时间
		length := 1000 << 20 //传输的数据大小
		totalRead := 0
		buf := make([]byte, 1024)
		readAll := func(r *bufio.Reader) (bytes []byte, err error) {
			defer func() {
				totalRead += len(bytes)
				if totalRead >= length {
					logs.Debugf("[读取]传输耗时: %0.1fMB/s\n", float64(totalRead/(1<<20))/time.Now().Sub(start).Seconds())
				}
			}()
			n, err := r.Read(buf)
			if err != nil {
				return nil, err
			}
			return buf[:n], nil
		}

		totalDeal := 0
		go listen.RunTCP(20145, func(s *server.Server) {
			s.Logger.SetLevel(common.LevelError)
			s.Logger.Enable(false)
			s.OnConnected(func(c *client.Client) {
				//c.SetBuffer(1024 * 10)
				c.OnReadFrom(readAll)
				c.OnDealMessage(func(c *client.Client, msg ios.Acker) {
					totalDeal += len(msg.Bytes())
					if totalDeal >= length {
						logs.Debugf("[处理]传输耗时: %0.1fMB/s\n", float64(totalDeal/(1<<20))/time.Now().Sub(start).Seconds())
						os.Exit(1)
					}
				})
			})

		})
		<-time.After(time.Second)
		<-redial.TCP("127.0.0.1:20145", func(c *client.Client) {
			c.Logger.Enable(false)
			c.Logger.SetLevel(common.LevelInfo)
			c.OnConnected(func(c *client.Client) error {
				data := make([]byte, length)
				start = time.Now()
				_, err := c.Write(data)
				logs.PrintErr(err)
				logs.Debugf("[发送]传输耗时: %0.1fMB/s\n", float64(length/(1<<20))/time.Now().Sub(start).Seconds())
				return nil
			})
			c.OnDealMessage(func(c *client.Client, msg ios.Acker) {
				logs.Debug(msg)
			})

		}).Done()

	}
}
