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

## ✨ 简介

`ios` 是一套面向 I/O 的统一运行时模型。

它不强调“内置了多少协议”，而强调“只要能抽象成 I/O，就能接入同一套运行模型”。

`module` 里只内置了一部分常见实现，比如 `TCP / UDP / SSH / WebSocket / Serial`，但这不是边界。

只要满足约定接口，就可以接入同一套流程：

- 📥 读取
- ✂️ 分包
- 📨 消息处理
- ⏱️ 超时控制
- 🔁 重连编排

适合需要长期维护连接、统一 I/O 逻辑的项目。

---

## 📦 安装

```bash
go get github.com/injoyai/ios/v2
```

---

## 🌟 为什么用它

很多 I/O 代码最后都会长成同一种样子：

- 建立连接
- 启动 goroutine
- 循环读取
- 手动处理超时
- 出错后重连
- 在不同协议之间重复写一遍

`ios` 的作用，是把这些重复结构整理成稳定的运行时模型。

它更适合这类场景：

- 你要长期维护连接，而不是一次性请求
- 你想统一 TCP、SSH、WebSocket、串口等接入方式
- 你希望把读取、分包、处理、超时、重连拆开管理
- 你不想在每种协议里都重复写一套底层循环

---

## 🔌 接入方式

真正的接入入口只有两个：

- 客户端实现 `ios.DialFunc`
- 服务端实现 `ios.ListenFunc`

客户端：

```go
type DialFunc func(ctx context.Context) (ReadWriteCloser, string, error)
```

服务端：

```go
type ListenFunc func() (Listener, error)

type Listener interface {
	io.Closer
	Accept() (ReadWriteCloser, string, error)
	Addr() string
}
```

最小约定就是：

- 🧩 客户端：提供 `DialFunc`
- 🧩 服务端：提供 `ListenFunc`
- 🔗 连接对象：返回 `ReadWriteCloser`

---

## 🚀 快速开始

下面是一个最小 TCP 客户端示例：

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

- 🔁 自动重连
- 📥 自定义读取策略
- 📨 消息处理回调
- 🤝 连接成功后的初始化逻辑

---

## 🖥️ 服务端示例

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

服务端同样围绕几件事展开：

- 接收连接
- 读取消息
- 处理消息
- 管理连接生命周期

---

## 🧠 理解模型

可以把 `ios` 理解成 4 层：

1. 建立连接：`Dial / Accept`
2. 定义读取：一条消息如何被切出来
3. 处理消息：业务如何消费数据
4. 管理生命周期：超时、关闭、重连、定时发送

这样代码会从“一个大循环里堆满读写逻辑”，变成“可替换、可组合、可维护的连接流”。

---

## 🎯 适用场景

- 持久 TCP 客户端
- 设备通信 / 工控接入
- SSH 控制通道
- WebSocket 消息处理
- 需要统一 `heartbeat / timeout / reconnect` 的项目
- 希望把底层 I/O 抽成稳定中间层的系统
