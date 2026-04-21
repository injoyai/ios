package ios

import (
	"io"
)

var _ AllReader = (*AllRead)(nil)

func NewAllReader(r Reader, f FReader) *AllRead {
	if v, ok := r.(*AllRead); ok {
		r = v.Reader
	}
	if f == nil {
		f = Buffer(make([]byte, DefaultBufferSize))
	}
	return &AllRead{
		Reader:     r,
		cache:      nil,
		fromReader: f,
	}
}

// AllRead ios.Reader转io.Reader
type AllRead struct {
	//只能是[Reader|BReader|AReader]类型
	Reader

	//用来缓存读取到的数据,方便下次使用
	//例如BReader,一次读取100字节,但是用户只取走40字节,剩下60字节缓存用于下次
	//不使用sync.Pool,因为大小不可知,防止被扩容造成的内存泄漏
	cache []byte

	//当Reader是io.Reader时有效
	fromReader FReader
}

func (this *AllRead) Read(p []byte) (n int, err error) {
	switch r := this.Reader.(type) {
	case BReader:
		if len(this.cache) == 0 {
			this.cache, err = r.ReadBytes()
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
			this.cache = a.Bytes()
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

	//一次性全部读取完,则清空缓冲区
	this.cache = nil
	return

}

func (this *AllRead) ReadBytes() (bs []byte, err error) {
	switch r := this.Reader.(type) {
	case BReader:
		return r.ReadBytes()

	case AReader:
		a, err := r.ReadAck()
		defer a.Ack()
		return a.Bytes(), err

	case io.Reader:
		return this.fromReader.ReadFrom(r)

	default:
		return nil, ErrUnknownReader

	}
}

func (this *AllRead) ReadAck() (Acker, error) {
	switch r := this.Reader.(type) {
	case BReader:
		bs, err := r.ReadBytes()
		if err != nil {
			return nil, err
		}
		return Ack(bs), nil

	case AReader:
		return r.ReadAck()

	case io.Reader:
		bs, err := this.fromReader.ReadFrom(this)
		if err != nil {
			return nil, err
		}
		return Ack(bs), err

	default:
		return nil, ErrUnknownReader

	}
}
