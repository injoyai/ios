package ios

import (
	"context"
	"io"
)

type (
	IO = AllReadWriteCloser

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

	AllReader interface {
		io.Reader
		MReader
		AReader
	}

	AllReadWriteCloser interface {
		io.ReadWriteCloser
		MReader
		AReader
	}

	// FReader FromReader 从io.Reader中读取数据
	FReader interface {
		ReadFrom(r io.Reader) ([]byte, error)
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

type FReadFunc func(r io.Reader) ([]byte, error)

func (this FReadFunc) ReadFrom(r io.Reader) ([]byte, error) { return this(r) }

type FRead2Func func(r Reader) ([]byte, error)

func (this FRead2Func) ReadFrom(r io.Reader) ([]byte, error) { return this(r) }

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
