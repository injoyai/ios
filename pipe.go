package ios

import (
	"io"
	"sync"
)

// Pipe 一个双向通道
func Pipe(cap int) (BReadWriteCloser, BReadWriteCloser) {
	return NewPiper(cap).Pipe()
}

func NewPiper(cap int) *Piper {
	return &Piper{
		Pipe1: newChan(cap),
		Pipe2: newChan(cap),
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

/*



 */

func newChan(cap int) *ChanIO {
	return &ChanIO{
		ch:        make(chan []byte, cap),
		closeSign: make(chan struct{}),
	}
}

type ChanIO struct {
	ch        chan []byte
	once      sync.Once
	closeSign chan struct{}
}

func (c *ChanIO) Write(p []byte) (int, error) {
	b := append([]byte(nil), p...)
	select {
	case <-c.closeSign:
		return 0, io.ErrClosedPipe
	case c.ch <- b:
		return len(b), nil
	}
}

func (c *ChanIO) ReadBytes() ([]byte, error) {
	select {
	case bs := <-c.ch:
		return bs, nil
	default:
	}

	select {
	case <-c.closeSign:
		return nil, io.EOF
	case bs := <-c.ch:
		return bs, nil
	}
}

func (c *ChanIO) Close() error {
	c.once.Do(func() {
		close(c.closeSign)
	})
	return nil
}
