package ios

import (
	"io"
	"time"

	"github.com/injoyai/base/chans"
)

// Pipe 一个双向通道
func Pipe(cap int, timeout ...time.Duration) (IO, IO) {
	return NewPiper(cap, timeout...).IO()
}

func NewPiper(cap int, timeout ...time.Duration) *Piper {
	return &Piper{
		Pipe1: chans.NewIO(cap, timeout...),
		Pipe2: chans.NewIO(cap, timeout...),
	}
}

type Piper struct {
	Pipe1 ReadWriteCloser
	Pipe2 ReadWriteCloser
}

func (this *Piper) Close() error {
	this.Pipe1.Close()
	this.Pipe2.Close()
	return nil
}

func (this *Piper) IO() (IO, IO) {
	i1 := struct {
		AllReader
		io.Writer
		io.Closer
	}{
		AllReader: NewAllReader(this.Pipe1, make(Bytes, DefaultBufferSize)),
		Writer:    this.Pipe2,
		Closer:    this,
	}
	i2 := struct {
		AllReader
		io.Writer
		io.Closer
	}{
		AllReader: NewAllReader(this.Pipe2, make(Bytes, DefaultBufferSize)),
		Writer:    this.Pipe1,
		Closer:    this,
	}
	return i1, i2
}
