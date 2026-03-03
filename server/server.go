package server

import (
	"context"
	"fmt"
	"iter"
	"sync"
	"time"

	"github.com/injoyai/base/maps/timeout"
	"github.com/injoyai/base/safe"
	"github.com/injoyai/ios/v2"
	"github.com/injoyai/ios/v2/client"
	"github.com/injoyai/ios/v2/module/common"
)

type Option func(s *Server)

func Run(listen ios.ListenFunc, op ...Option) error {
	return RunContext(context.Background(), listen, op...)
}

func RunContext(ctx context.Context, listen ios.ListenFunc, op ...Option) error {
	s, err := New(listen, op...)
	if err != nil {
		return err
	}
	return s.Run(ctx)
}

func New(listen ios.ListenFunc, op ...Option) (*Server, error) {
	listener, err := listen()
	if err != nil {
		return nil, err
	}
	s := &Server{
		Event:    &Event{},
		Closer:   safe.NewCloser(),
		Runner2:  safe.NewRunner2(nil),
		Logger:   common.NewLogger(),
		listener: listener,
		timeout: timeout.New().SetDealFunc(func(key interface{}) error {
			return key.(*client.Client).CloseWithErr(ios.ErrWithTimeout)
		}),
		client: make(map[string]*client.Client),
	}
	s.Runner2.SetFunc(s.run)
	s.timeout.SetTimeout(time.Minute * 3) //3分钟超时(3-检查间隔会超时)
	s.timeout.SetInterval(time.Minute)    //1分钟检查一次
	s.Closer.SetCloseFunc(func(err error) error {
		//关闭全部客户端,是否关闭?,net包是不关闭已连接的客户端,可以方便热启动
		//s.CloseAllClient(err)
		//服务关闭事件
		s.Logger.Infof("[%s] 关闭服务...\n", listener.Addr())
		if s.Event != nil && s.Event.onClose != nil {
			s.Event.onClose(s, err)
		}
		return listener.Close()
	})
	for _, v := range op {
		v(s)
	}
	//放在用户选项之后,方便用户控制是否输出
	s.Logger.Infof("[%s] 开启服务成功...\n", listener.Addr())
	if s.Event.onOpen != nil {
		s.Event.onOpen(s)
	}
	return s, nil
}

type Server struct {
	*Event                      //事件
	*safe.Closer                //closer
	*safe.Runner2               //runner
	Logger        common.Logger //日志

	listener ios.Listener     //listene
	timeout  *timeout.Timeout //超时机制

	client   map[string]*client.Client //客户端
	clientMu sync.RWMutex              //锁
}

// Timer 定时器
func (this *Server) Timer(t time.Duration, f Option) {
	tick := time.NewTicker(t)
	defer tick.Stop()
	for {
		select {
		case <-this.Closer.Done():
			return
		case _, ok := <-tick.C:
			if ok {
				f(this)
			}
		}
	}
}

// GetClient 获取客户端
func (this *Server) GetClient(key string) *client.Client {
	this.clientMu.RLock()
	defer this.clientMu.RUnlock()
	return this.client[key]
}

// GetClientLen 获取客户端数量
func (this *Server) GetClientLen() int {
	this.clientMu.RLock()
	defer this.clientMu.RUnlock()
	return len(this.client)
}

// RangeClient 遍历客户端
func (this *Server) RangeClient(f func(k string, c *client.Client) bool) {
	for k, c := range this.Clients() {
		if !f(k, c) {
			return
		}
	}
}

// Clients 遍历客户端
func (this *Server) Clients() iter.Seq2[string, *client.Client] {
	this.clientMu.RLock()
	defer this.clientMu.RUnlock()
	return func(yield func(string, *client.Client) bool) {
		for k, c := range this.client {
			if !yield(k, c) {
				return
			}
		}
	}
}

// CloseClient 关闭客户端
func (this *Server) CloseClient(key string, err error) {
	c := this.GetClient(key)
	if c != nil {
		c.CloseWithErr(err)
	}
	this.clientMu.Lock()
	defer this.clientMu.Unlock()
	delete(this.client, key)
}

// CloseAllClient 关闭全部客户端
func (this *Server) CloseAllClient(err error) {
	for _, c := range this.Clients() {
		c.CloseWithErr(err)
	}
	this.clientMu.Lock()
	defer this.clientMu.Unlock()
	this.client = make(map[string]*client.Client)
}

func (this *Server) run(ctx context.Context) error {
	for {
		c, k, err := this.listener.Accept()
		if err != nil {
			return err
		}

		go func(ctx context.Context, k string, c ios.ReadWriteCloser) {
			cli := client.New(nil)
			cli.Logger = this.Logger
			cli.SetOption(this.Event.clientOptions...)
			cli.SetReadWriteCloser(k, c)

			//触发服务端连接/断开事件
			this.Logger.Infof("[%s] 新的客户端连接...\n", cli.Key())
			defer func() {
				//客户端断开连接事件
				if this.Event.onDisConnected != nil {
					this.Event.onDisConnected(this, cli, cli.Err())
				}
				this.Logger.Infof("[%s] 客户端断开连接...\n", cli.Key())
			}()

			//如果客户端被关闭,则退出
			if cli.Closed() {
				return
			}

			//取消重试,客户端是被连接
			cli.SetRedial(false)
			//取消读取超时机制,取消客户端,在服务端实现
			cli.SetReadTimeout(0)

			//设置修改key事件
			cli.OnKeyChange(func(c *client.Client, oldKey string) {
				this.onChangeKey(c, oldKey)
			})

			//保持读超时状态
			cli.OnDealMessage(func(c *client.Client, message ios.Acker) {
				this.timeout.Keep(c)
			})

			//保持写超时状态
			cli.OnWriteWith(func(bs []byte) ([]byte, error) {
				this.timeout.Keep(c)
				return bs, nil
			})

			//把客户端设置到缓存
			this.clientMu.Lock()
			this.client[cli.Key()] = cli
			this.clientMu.Unlock()

			//运行客户端
			_ = cli.Run(ctx)

			//等待结束之后从缓存删除客户端
			this.clientMu.Lock()
			delete(this.client, cli.Key())
			this.clientMu.Unlock()

		}(ctx, k, c)
	}
}

func (this *Server) onChangeKey(c *client.Client, oldKey string) {
	//判断是否存在老连接,存在则关闭老连接(被挤下线)
	this.clientMu.RLock()
	//查找这个新key是否存在实例,存在则判断2个是否是一个,不是一个则关闭老的那个
	old, ok := this.client[c.Key()]
	this.clientMu.RUnlock()
	if ok && old != c {
		old.CloseWithErr(fmt.Errorf("重复标识(%s),关闭老客户端", old.Key()))
	}
	//保存到缓存中
	this.clientMu.Lock()
	delete(this.client, oldKey)
	this.client[c.Key()] = c
	this.clientMu.Unlock()
}
