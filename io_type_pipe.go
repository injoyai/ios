package ios

import (
	"github.com/injoyai/base/chans"
	"time"
)

// Pipe 一个双向通道
func Pipe(cap uint, timeout ...time.Duration) (IO, IO) {
	return NewPiper(cap, timeout...).IO()
}

func MPipe(cap uint, timeout ...time.Duration) (MIO, MIO) {
	return NewPiper(cap, timeout...).MIO()
}

func APipe(cap uint, timeout ...time.Duration) (AIO, AIO) {
	return NewPiper(cap, timeout...).AIO()
}

func NewPiper(cap uint, timeout ...time.Duration) *Piper {
	return &Piper{
		Pipe1: chans.NewIO(cap, timeout...),
		Pipe2: chans.NewIO(cap, timeout...),
	}
}

type Piper struct {
	Pipe1 *chans.IO
	Pipe2 *chans.IO
}

func (this *Piper) Close() error {
	this.Pipe1.Close()
	this.Pipe2.Close()
	return nil
}

func (this *Piper) IO() (IO, IO) {
	return NewIO(this.Pipe1, this.Pipe2, this),
		NewIO(this.Pipe2, this.Pipe1, this)
}

func (this *Piper) MIO() (MIO, MIO) {
	return NewMIO(this.Pipe1, this.Pipe2, this),
		NewMIO(this.Pipe2, this.Pipe1, this)
}

func (this *Piper) AIO() (AIO, AIO) {
	return NewAIO(M2AReader(this.Pipe1), this.Pipe2, this),
		NewAIO(M2AReader(this.Pipe2), this.Pipe1, this)
}
