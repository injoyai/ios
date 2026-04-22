package ios

import (
	"bufio"
	"io"
)

const (
	DefaultBufferSize = 1024
)

type Bytes []byte

func (this Bytes) ReadFrame(r *bufio.Reader) ([]byte, error) {
	return this.ReadReader(r)
}

func (this Bytes) ReadBuffer(r *bufio.Reader) ([]byte, error) {
	return this.ReadReader(r)
}

func (this Bytes) ReadReader(r io.Reader) ([]byte, error) {
	n, err := r.Read(this)
	if err != nil {
		return nil, err
	}
	return this[:n], nil
}

func NewBuffer(r io.Reader, buf []byte) *Buffer {
	if buf == nil {
		buf = make([]byte, DefaultBufferSize)
	}
	return &Buffer{
		Reader: r,
		buf:    buf,
	}
}

/*
Buffer
缓存读取,用来替代bufio.Reader,原因是不可控
这个能自定义buf,方便内存复用
*/
type Buffer struct {
	// 原始Reader,注意使用安全
	io.Reader

	//用来缓存读取到的数据,方便下次使用
	//例如BReader,一次读取100字节,但是用户只取走40字节,剩下60字节缓存用于下次
	buf []byte

	//数据位置下标
	i, j int
}

func (this *Buffer) Cap() int {
	return cap(this.buf)
}

// Len 返回已缓存数据长度
func (this *Buffer) Len() int {
	return this.j - this.i
}

func (this *Buffer) Reset(r io.Reader) {
	this.Reader = r
	this.i = 0
	this.j = 0
}

func (this *Buffer) Clear() {
	this.Reset(nil)
}

func (this *Buffer) Read(p []byte) (int, error) {

	if this.j <= this.i {
		//从底层IO读取数据到缓存
		n, err := this.Reader.Read(this.buf)
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
