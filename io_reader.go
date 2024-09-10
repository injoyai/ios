package ios

import (
	"io"
)

type ReadOption func(p []byte, err error)

func NewAllReader(r Reader, f Read, op ...ReadOption) AllReader {
	return &AllRead{
		Reader:  r,
		Handler: f,
		Options: op,
	}
}

// AllRead ios.Reader转io.Reader
type AllRead struct {
	//只能是[Reader|MReader|AReader]类型
	Reader
	//当Reader是io.Reader时有效
	Handler Read
	cache   []byte
	Options []ReadOption
}

func (this *AllRead) Read(p []byte) (n int, err error) {
	defer func() {
		for _, f := range this.Options {
			if f != nil {
				f(p[:n], err)
			}
		}
	}()

	switch r := this.Reader.(type) {
	case MReader:
		if len(this.cache) == 0 {
			this.cache, err = r.ReadMessage()
			if err != nil {
				return
			}
		}
	case AReader:
		if len(this.cache) == 0 {
			a, err := r.ReadAck()
			if err != nil {
				return 0, err
			}
			this.cache = a.Payload()
		}

	case io.Reader:
		return r.Read(p)

	default:
		return 0, ErrUnknownReader

	}

	//从缓存(上次剩余的字节)复制数据到p
	n = copy(p, this.cache)
	if n < len(this.cache) {
		this.cache = this.cache[n:]
		return
	}

	//一次性全部读取完
	this.cache = nil
	return

}

func (this *AllRead) ReadMessage() (bs []byte, err error) {
	defer func() {
		for _, f := range this.Options {
			if f != nil {
				f(bs, err)
			}
		}
	}()

	switch r := this.Reader.(type) {
	case MReader:
		return r.ReadMessage()
	case AReader:
		a, err := r.ReadAck()
		defer a.Ack()
		return a.Payload(), err
	default:
		if this.Handler == nil {
			this.Handler = NewRead4KB()
		}
		return this.Handler(this)
	}
}

func (this *AllRead) ReadAck() (a Acker, err error) {
	defer func() {
		for _, f := range this.Options {
			if f != nil {
				f(a.Payload(), err)
			}
		}
	}()

	switch r := this.Reader.(type) {
	case MReader:
		bs, err := r.ReadMessage()
		return Ack(bs), err
	case AReader:
		return r.ReadAck()
	default:
		if this.Handler == nil {
			this.Handler = NewRead4KB()
		}
		bs, err := this.Handler(this)
		return Ack(bs), err
	}
}

type IOer struct {
	*AllRead
	io.Writer
	io.Closer
}
