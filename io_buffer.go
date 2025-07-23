package ios

import (
	"io"
)

const (
	DefaultBufferSize = 4096
)

func NewBufferReader(r io.Reader, buf []byte) *BufferReader {
	if buf == nil {
		buf = make([]byte, DefaultBufferSize)
	}
	return &BufferReader{
		r:   r,
		buf: buf,
	}
}

/*
BufferReader
缓存读取,用来替代bufio.Reader,原因是不可控
这个能自定义buf,方便内存复用
*/
type BufferReader struct {
	// reader provided by the client
	r io.Reader

	//用来缓存读取到的数据,方便下次使用
	//例如MReader,一次读取100字节,但是用户只取走40字节,剩下60字节缓存用于下次
	buf []byte

	//数据位置下标
	i, j int
}

func (this *BufferReader) Cap() int {
	return cap(this.buf)
}

// Len 返回已缓存数据长度
func (this *BufferReader) Len() int {
	return this.j - this.i
}

func (this *BufferReader) SetReader(r io.Reader) {
	this.r = r
}

func (this *BufferReader) Reset() {
	this.i = 0
	this.j = 0
}

func (this *BufferReader) Read(p []byte) (int, error) {

	if this.j <= this.i {
		//从底层IO读取数据到缓存
		n, err := this.r.Read(this.buf)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, io.ErrNoProgress
		}
		this.i = 0
		this.j = n
	}

	//从缓存(上次剩余的字节)复制数据到p
	n := copy(p, this.buf[this.i:this.j])
	this.i += n
	return n, nil
}
