package ios

import (
	"context"
	"io"
)

type (
	Reader interface {
		//Reader为这三种类型 [io.Reader|AReader|MReader] 如何用泛型实现?
	}

	ReadCloser interface {
		Reader
		io.Closer
	}

	ReadeWriteCloser interface {
		Reader
		io.WriteCloser
	}

	Closer interface {
		io.Closer
		Closed() bool
	}

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
		io.Closer
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

// WriteFunc 写入函数
type WriteFunc func(p []byte) (int, error)

func (this WriteFunc) Write(p []byte) (int, error) { return this(p) }

// CloseFunc 关闭函数
type CloseFunc func() error

func (this CloseFunc) Close() error { return this() }

type Ack []byte

func (this Ack) Ack() error { return nil }

func (this Ack) Payload() []byte { return this }

type DialFunc func(ctx context.Context) (ReadeWriteCloser, string, error)
