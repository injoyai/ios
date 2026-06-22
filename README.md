# ios

<p align="center">
  <b>连接万物，万物皆 I/O</b>
</p>

<p align="center">
  用统一模型组织连接、读取、消息处理、超时与重连；只要满足接口，就可以接入。
</p>

<p align="center">
  <img alt="Go Version" src="https://img.shields.io/badge/Go-1.20%2B-111111?style=flat-square&logo=go&logoColor=5cc8ff">
  <img alt="Version" src="https://img.shields.io/badge/module-v2-111111?style=flat-square&logo=go&logoColor=f4f4f4">
  <img alt="Focus" src="https://img.shields.io/badge/focus-connection%20runtime-111111?style=flat-square&logo=buffer&logoColor=c7f36b">
</p>

---

## 简介

`ios` 不是只服务于某几种协议的连接库，而是一套面向 I/O 的统一运行时模型。

它的目标很直接：**连接万物，万物皆 I/O**。

`module` 目录里只实现了一部分常见接入方式，比如 `TCP / UDP / SSH / WebSocket / Serial / Bridge`，但它并不把边界限定在这些模块上。

只要你的对象满足约定接口，就可以被接入同一套处理流程：

- 用什么方式读取数据
- 一条消息如何切分
- 收到消息后交给谁处理
- 超时、心跳、重连如何组织
- 不同连接形态如何复用同一套处理方式

适合需要长期维护连接、统一 I/O 逻辑、降低接入成本的项目。

---

## 能力概览

| 能力 | 说明 |
| --- | --- |
| `Client` | 统一的客户端连接抽象 |
| `Server` | TCP / UDP 等服务端接入能力 |
| `Read` / `Ack` | 面向消息边界的读取与确认模型 |
| `Redial` | 自动重连与重拨编排 |
| `Timeout` | 读超时控制，兼容不同连接能力 |
| `Bridge` | 双向桥接两个读写端 |
| `Hooks` | `OnConnected`、`OnDealMessage`、`OnReadFrom` 等扩展点 |

---

## 安装

```bash
go get github.com/injoyai/ios/v2
```

---

## 为什么用它

很多 I/O 代码最后都会长成同一种样子：

- 建立连接
- 开 goroutine
- 循环读取
- 手动处理超时
- 出错后重连
- 在不同协议之间重复写一遍

`ios` 的作用，是把这些重复结构整理成稳定的运行时模型。

它不依赖 `module` 里是否已经内置某种实现；只要满足接口约定，就能接入这套组织方式。

它提供的不是某一种协议能力，而是一种更统一的组织方式：

- **连接是对象**：可以运行、关闭、重连、观测
- **读取是策略**：按行、按长度、按分隔符、按块读取都可替换
- **消息有边界**：业务处理的是消息，而不是裸字节流
- **生命周期可编排**：连接、心跳、超时、断线、重拨都能挂入流程

---

## 接入接口

真正决定“能不能接入 `ios`”的入口有两个：

- 客户端实现 `ios.DialFunc`
- 服务端实现 `ios.ListenFunc`

### 客户端接入

客户端只需要提供一个拨号函数：

```go
type DialFunc func(ctx context.Context) (ReadWriteCloser, string, error)
```

也就是说，你需要返回：

- 一个 `ReadWriteCloser`
- 一个连接标识 `string`
- 一个 `error`

常见理解可以是：

- `ReadWriteCloser`：真正的 I/O 对象
- `string`：这个连接的 key，通常是地址、设备号或自定义标识
- `error`：拨号失败时返回

只要你能封装出一个 `DialFunc`，这个客户端就可以被 `ios/client` 接管。

### 服务端接入

服务端只需要提供一个监听函数：

```go
type ListenFunc func() (Listener, error)
```

其中 `Listener` 需要满足：

```go
type Listener interface {
	io.Closer
	Accept() (ReadWriteCloser, string, error)
	Addr() string
}
```

也就是说，服务端接入的关键不是某个固定协议，而是你能否提供一个可 `Accept()` 的监听对象。

每次接收到新连接时，返回：

- 一个 `ReadWriteCloser`
- 一个连接标识 `string`
- 一个 `error`

只要你能封装出一个 `ListenFunc`，这个服务端就可以被 `ios/server` 接管。

### 最小约定

所以从接入角度看，`module` 目录里的实现只是示例，不是边界。

真正的最小约定是：

- **客户端**：实现 `DialFunc`
- **服务端**：实现 `ListenFunc`
- **连接对象**：返回值里提供 `ReadWriteCloser`

这也是这个库“连接万物，万物皆 I/O”的核心前提。
---

## 快速开始

下面是一个最小 TCP 长连接客户端示例：

```go
package main

import (
	"time"

	"github.com/injoyai/ios/v2"
	"github.com/injoyai/ios/v2/client"
	"github.com/injoyai/ios/v2/client/dial"
	"github.com/injoyai/ios/v2/client/redial"
)

func main() {
	addr := "127.0.0.1:10086"

	redial.Run(dial.WithTCP(addr), func(c *client.Client) {
		c.Logger.Debug()
		c.Logger.WithUTF8()

		c.OnReadFrom(ios.NewRead4KB())
		c.OnDealMessage(func(c *client.Client, msg ios.Acker) {
			// 处理收到的数据
		})

		c.OnConnected(func(c *client.Client) error {
			c.GoTimerWriter(time.Minute, func(w ios.MoreWriter) error {
				_, err := w.WriteString("心跳")
				return err
			})
			return nil
		})
	})
}
```

这个示例包含：

- 自动重连
- 自定义读取策略
- 消息处理回调
- 连接成功后的初始化逻辑
- 定时心跳发送

---

## 常见场景

### TCP

适合长连接客户端、设备通信、网关通信、二进制协议传输。

### SSH

把 SSH 会话接入统一消息处理模型：

```go
package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/injoyai/ios/v2"
	"github.com/injoyai/ios/v2/client"
	"github.com/injoyai/ios/v2/client/redial"
	"github.com/injoyai/ios/v2/module/ssh"
)

func main() {
	c := redial.SSH(&ssh.Config{
		Address:  os.Args[1],
		User:     os.Args[2],
		Password: os.Args[3],
	})

	c.Logger.Debug(false)
	c.OnDealMessage(func(c *client.Client, msg ios.Acker) {
		fmt.Print(string(msg.Payload()))
	})

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

### WebSocket

适合实时消息通道或桥接型场景：

```go
package main

import (
	"github.com/injoyai/ios/v2"
	"github.com/injoyai/ios/v2/client"
	"github.com/injoyai/ios/v2/client/redial"
)

func main() {
	redial.RunWebsocket("http://127.0.0.1:80/ws", nil, func(c *client.Client) {
		c.Logger.Debug()
		c.Logger.WithUTF8()
		c.OnDealMessage(func(c *client.Client, msg ios.Acker) {
			// 处理读取到的数据
		})
	})
}
```

---

## 服务端

`ios` 也提供统一的服务端连接处理模型：

```go
package main

import (
	"github.com/injoyai/ios/v2"
	"github.com/injoyai/ios/v2/server"
)

func main() {
	s := server.NewTCP(":10086")
	s.OnReadFrom(ios.NewRead4KB())
	s.OnDealMessage(func(c ios.ReadWriteCloser, msg ios.Acker) {
		// 处理客户端消息
	})
	s.Run()
}
```

---

## 心智模型

可以把 `ios` 理解成 4 层：

1. 建立连接：`Dial / Accept`
2. 定义读取：一条消息如何被切出来
3. 处理消息：业务如何消费数据
4. 管理生命周期：超时、关闭、重连、定时发送

这样代码会从“一个大循环里堆满读写逻辑”，变成“可替换、可组合、可维护的连接流”。

---

## 适用项目

- 设备通信 / 工控接入
- 持久 TCP 客户端
- SSH 控制通道
- WebSocket 消息处理
- 需要统一 heartbeat / timeout / reconnect 的项目
- 希望把底层 I/O 抽成稳定中间层的系统

---

## 建议阅读顺序

如果你第一次接入，建议按这个顺序看：

1. `client`
2. `server`
3. `module`
4. 一个最小 TCP 示例
5. 再补心跳、超时和重连
