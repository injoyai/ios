package ios

import (
	"context"
	"io"
)

type (
	IO  = io.ReadWriteCloser
	MIO = MReadWriteCloser
	AIO = AReadWriteCloser

	Reader interface {
		//Reader为这三种类型 [io.Reader|AReader|MReader] 如何用泛型实现?
	}

	ReadCloser interface {
		Reader
		io.Closer
	}

	ReadWriteCloser interface {
		Reader
		io.WriteCloser
	}

	Closer interface {
		io.Closer
		Closed() bool
	}

	// AReader 更加兼容各种协议,例如MQTT,RabbitMQ等
	AReader interface {
		ReadAck() (Acker, error)
	}

	AReadCloser interface {
		AReader
		io.Closer
	}

	AReadWriter interface {
		AReader
		io.Writer
	}

	AReadWriteCloser interface {
		AReader
		io.Writer
		io.Closer
	}

	// MReader 使用更方便,就是分包后的IO
	MReader interface {
		ReadMessage() ([]byte, error)
	}

	MReadWriter interface {
		MReader
		io.Writer
	}

	MReadCloser interface {
		MReader
		io.Closer
	}

	MReadWriteCloser interface {
		MReader
		io.Writer
		io.Closer
	}

	Runner interface {
		Closer
		Run() error
		Running() bool
	}

	Base64Writer interface {
		WriteBase64(s string) error
	}

	HEXWriter interface {
		WriteHEX(s string) error
	}

	JsonWriter interface {
		WriteJson(a any) error
	}

	AnyWriter interface {
		WriteAny(a any) error
	}

	ChanWriter interface {
		WriteChan(c chan any) error
	}

	// MoreWriter 各种方式的写入
	MoreWriter interface {
		io.Writer
		io.StringWriter
		io.ByteWriter
		Base64Writer
		HEXWriter
		JsonWriter
		AnyWriter
		ChanWriter
	}

	Listener interface {
		io.Closer
		Accept() (ReadWriteCloser, string, error)
		Addr() string
	}
)

// Acker 兼容MQ等需要确认的场景
type Acker interface {
	Payload() []byte
	Ack() error
}

//=================================Func=================================

// ReadFunc 读取函数
type ReadFunc func(p []byte) (int, error)

func (this ReadFunc) Read(p []byte) (int, error) { return this(p) }

type AReadFunc func() (Acker, error)

func (this AReadFunc) ReadAck() (Acker, error) { return this() }

type MReadFunc func() ([]byte, error)

func (this MReadFunc) ReadMessage() ([]byte, error) { return this() }

// WriteFunc 写入函数
type WriteFunc func(p []byte) (int, error)

func (this WriteFunc) Write(p []byte) (int, error) { return this(p) }

// CloseFunc 关闭函数
type CloseFunc func() error

func (this CloseFunc) Close() error { return this() }

type Ack []byte

func (this Ack) Ack() error { return nil }

func (this Ack) Payload() []byte { return this }

type DialFunc func(ctx context.Context) (ReadWriteCloser, string, error)

type ListenFunc func() (Listener, error)

type WriteTo func(w io.Writer) error

//type ReadFrom func(r Reader) ([]byte, error)

type Read func(r io.Reader) ([]byte, error)

//=================================Struct=================================

type ToMReader struct {
	io.Reader
	Handler Read
}

func (this *ToMReader) ReadMessage() ([]byte, error) {
	return this.Handler(this.Reader)
}

type ToAReader struct {
	io.Reader
	Handler Read
}

func (this *ToAReader) ReadAck() (Acker, error) {
	bs, err := this.Handler(this.Reader)
	if err != nil {
		return nil, err
	}
	return Ack(bs), nil
}

type ToReader struct {
	MReader
	readCache []byte
}

func (this *ToReader) Read(p []byte) (n int, err error) {

	if len(this.readCache) == 0 {
		this.readCache, err = this.ReadMessage()
		if err != nil {
			return
		}
	}

	//从缓存(上次剩余的字节)复制数据到p
	n = copy(p, this.readCache)
	if n < len(this.readCache) {
		this.readCache = this.readCache[n:]
		return
	}

	this.readCache = nil
	return
}

type ToIO struct {
	io.Writer
	io.Closer
	ToReader
}

type ToMIO struct {
	io.Writer
	io.Closer
	ToMReader
}

type ToAIO struct {
	io.Writer
	io.Closer
	ToAReader
}
