# io


### 如何使用

#### 如何连接TCP

```go

package main

import (
	"github.com/injoyai/ios"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/ios/client/redial"
	"time"
)

func main() {
	addr := "127.0.0.1:10086"
	redial.Run(dial.WithTCP(addr),
		func(c *client.Client) {
			c.Logger.Debug()                      //开启打印日志
			c.Logger.WithUTF8()                   //打印日志编码ASCII
			c.Event.OnReadFrom = ios.NewRead4KB() //设置读取方式,一次读取4KB
			c.Event.OnDealMessage = func(c *client.Client, msg ios.Acker) {
				// todo 业务逻辑,处理读取到的数据
			}
			c.Event.OnConnected = func(c *client.Client) error {
				c.GoTimerWriter(time.Minute, func(w ios.MoreWriter) error {
					_, err := w.WriteString("心跳") //定时发送心跳
					return err
				})
                return nil
            }
		}, 
    )
}

```

#### 如何连接SSH

```go

package main

import (
	"bufio"
	"fmt"
	"github.com/injoyai/ios/client/redial"
	"github.com/injoyai/ios/module/ssh"
	"os"
)

func main() {
	c := redial.SSH(&ssh.Config{
		Address:  os.Args[1],
		User:     os.Args[2],
		Password: os.Args[3],
	})
	c.Logger.Debug(false)
	c.Event.OnDealMessage = func(c *client.Client, msg ios.Acker) {
		fmt.Print(string(msg.Payload()))
	}
	go c.Run()
	reader := bufio.NewReader(os.Stdin)
	for {
		select {
		case <-c.Done():
			return
		default:
			bs, _, _ := reader.ReadLine()
			c.Write(append(bs, '\n'))
		}
	}
}

```

#### 如何连接Websocket

```go
package main

import(
	"bufio"
	"github.com/injoyai/ios"
	"github.com/injoyai/ios/client/redial"
	"os"
)

func main(){
	redial.RunWebsocket("http://127.0.0.1:80/ws",nil,
		func(c *client.Client) {
			c.Logger.Debug()
			c.Logger.WithUTF8()
			c.Event.OnDealMessage= func(c *client.Client, msg ios.Acker){
				// todo 业务逻辑,处理读取到的数据
			}
        })
	
}

```