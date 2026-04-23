package ios

import (
	"bufio"
	"context"
	"io"
	"time"
)

type (
	Reader interface {
		//Reader为这三种类型 [io.Reader|AReader|BReader] 如何用泛型实现?
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

	// BReader 使用更方便,就是分包后的IO
	BReader interface {
		ReadBytes() ([]byte, error)
	}

	BReadWriter interface {
		BReader
		io.Writer
	}

	BReadCloser interface {
		BReader
		io.Closer
	}

	BReadWriteCloser interface {
		BReader
		io.Writer
		io.Closer
	}

	AllReader interface {
		AReader
		BReader
		io.Reader
	}

	AllReadWriteCloser interface {
		AReader
		BReader
		io.ReadWriteCloser
	}

	// FReader FromReader 从*bufio.Reader中读取数据
	FReader interface {
		ReadFrom(r *bufio.Reader) ([]byte, error)
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

	// Acker 兼容MQ等需要确认的场景
	Acker interface {
		Bytes() []byte
		Ack() error
	}

	SetReadDeadliner interface {
		SetReadDeadline(t time.Time) error
	}
)

//=================================Func=================================

type ReadFunc func(p []byte) (int, error)

func (this ReadFunc) Read(p []byte) (int, error) { return this(p) }

type AReadFunc func() (Acker, error)

func (this AReadFunc) ReadAck() (Acker, error) { return this() }

type BReadFunc func() ([]byte, error)

func (this BReadFunc) ReadBytes() ([]byte, error) { return this() }

type FReadFunc func(r *bufio.Reader) ([]byte, error)

func (this FReadFunc) ReadFrom(r *bufio.Reader) ([]byte, error) { return this(r) }

type WriteFunc func(p []byte) (int, error)

func (this WriteFunc) Write(p []byte) (int, error) { return this(p) }

type CloseFunc func() error

func (this CloseFunc) Close() error { return this() }

type Ack []byte

func (this Ack) Ack() error { return nil }

func (this Ack) Bytes() []byte { return this }

type DialFunc func(ctx context.Context) (ReadWriteCloser, string, error)

type ListenFunc func() (Listener, error)
