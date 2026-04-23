package ios

import (
	"io"

	"github.com/injoyai/base/chans"
)

// Pipe 一个双向通道
func Pipe(cap int) (BReadWriteCloser, BReadWriteCloser) {
	return NewPiper(cap).Pipe()
}

func NewPiper(cap int) *Piper {
	return &Piper{
		Pipe1: chans.NewIO(cap),
		Pipe2: chans.NewIO(cap),
	}
}

type Piper struct {
	Pipe1 BReadWriteCloser
	Pipe2 BReadWriteCloser
}

func (this *Piper) Close() error {
	this.Pipe1.Close()
	this.Pipe2.Close()
	return nil
}

func (this *Piper) Pipe() (BReadWriteCloser, BReadWriteCloser) {
	i1 := struct {
		BReader
		io.Writer
		io.Closer
	}{
		BReader: this.Pipe1,
		Writer:  this.Pipe2,
		Closer:  this,
	}
	i2 := struct {
		BReader
		io.Writer
		io.Closer
	}{
		BReader: this.Pipe2,
		Writer:  this.Pipe1,
		Closer:  this,
	}
	return i1, i2
}
