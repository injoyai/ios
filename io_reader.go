package ios

import (
	"io"
)

func NewFreeFReader(f func(r io.Reader) ([]byte, error)) FreeFReader {
	if f == nil {
		return nil
	}
	return FreeFReadFunc(f)
}

type FreeFReadFunc func(r io.Reader) ([]byte, error)

func (this FreeFReadFunc) ReadFrom(r io.Reader) ([]byte, error) {
	return this(r)
}

func (this FreeFReadFunc) Free() {}

func NewAllReader(r Reader, f FreeFReader) *AllRead {
	if f == nil {
		f = DefaultFReaderPool.Get()
	}
	if v, ok := r.(*AllRead); ok {
		v.freeFromReader = f
		return v
	}
	return &AllRead{
		Reader:         r,
		freeFromReader: f,
	}
}

// AllRead ios.Reader转io.Reader
type AllRead struct {
	//只能是[Reader|MReader|AReader]类型
	Reader

	//用来缓存读取到的数据,方便下次使用
	//例如MReader,一次读取100字节,但是用户只取走40字节,剩下60字节缓存用于下次
	//不使用sync.Pool,因为大小不可知,防止被扩容造成的内存泄漏
	cache []byte

	//当Reader是io.Reader时有效,带Free(用于内存释放)的FromReader
	//替换的时候,推荐手动Free(),能回到pool中,否则按正常流程被GC()
	freeFromReader FreeFReader
}

func (this *AllRead) Free() {
	this.Reader = nil
	this.cache = nil
	this.freeFromReader.Free()
}

func (this *AllRead) Read(p []byte) (n int, err error) {
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

	//一次性全部读取完,则清空缓冲区
	this.cache = this.cache[:0]
	return

}

func (this *AllRead) ReadMessage() (bs []byte, err error) {
	switch r := this.Reader.(type) {
	case MReader:
		return r.ReadMessage()
	case AReader:
		a, err := r.ReadAck()
		defer a.Ack()
		return a.Payload(), err
	case io.Reader:
		return this.freeFromReader.ReadFrom(r)
	default:
		return nil, ErrUnknownReader
	}
}

func (this *AllRead) ReadAck() (Acker, error) {
	switch r := this.Reader.(type) {
	case MReader:
		bs, err := r.ReadMessage()
		if err != nil {
			return nil, err
		}
		return Ack(bs), nil
	case AReader:
		return r.ReadAck()
	case io.Reader:
		bs, err := this.freeFromReader.ReadFrom(this)
		if err != nil {
			return nil, err
		}
		return Ack(bs), err
	default:
		return nil, ErrUnknownReader
	}
}

type IOer struct {
	*AllRead
	io.Writer
	io.Closer
}
