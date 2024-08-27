package ios

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

func NewRead(buf []byte) func(r io.Reader) ([]byte, error) {
	if buf == nil {
		buf = make([]byte, 1024*4)
	}
	return func(r io.Reader) ([]byte, error) {
		n, err := r.Read(buf)
		if err != nil {
			return nil, err
		}
		return buf[:n], nil
	}
}

// NewReadLeast 新建读取函数,至少读取设置的字节
func NewReadLeast(least int) func(r io.Reader) ([]byte, error) {
	buf := make([]byte, least)
	return func(r io.Reader) ([]byte, error) {
		_, err := io.ReadAtLeast(r, buf, least)
		return buf, err
	}
}

// NewReadKB 新建读取函数,按KB读取
func NewReadKB(n int) func(r io.Reader) ([]byte, error) {
	return NewRead(make([]byte, 1024*n))
}

// NewReadMost 新建读取函数,按最大字节数读取
func NewReadMost(max int) func(r io.Reader) ([]byte, error) {
	return NewRead(make([]byte, max))
}

// NewReadFromWithRead 读取函数
func NewReadFromWithRead(f Read) func(r Reader) ([]byte, error) {
	if f == nil {
		f = NewRead(make([]byte, 1024*4))
	}
	var buffer *bufio.Reader
	return func(r Reader) ([]byte, error) {
		switch v := r.(type) {
		case MReader:
			return v.ReadMessage()

		case AReader:
			a, err := v.ReadAck()
			if err != nil {
				return nil, err
			}
			defer a.Ack()
			return a.Payload(), nil

		case *bufio.Reader:
			return f(v)

		case io.Reader:
			if buffer == nil {
				buffer = bufio.NewReaderSize(v, 1024*4)
			}
			return f(buffer)

		default:
			return nil, fmt.Errorf("未知类型: %T, 未实现[Reader|MReader|AReader]", r)

		}

	}
}

// NewReadFrom 读取函数
func NewReadFrom(buf []byte) func(r Reader) ([]byte, error) {
	return NewReadFromWithRead(NewRead(buf))
}

type Bytes []byte

func (this Bytes) ReadFrom(r io.Reader) ([]byte, error) {
	n, err := r.Read(this)
	if err != nil {
		return nil, err
	}
	return this[:n], nil
}

// ReadByte 读取一字节
func ReadByte(r io.Reader) (byte, error) {
	switch v := r.(type) {
	case io.ByteReader:
		return v.ReadByte()
	default:
		b := make([]byte, 1)
		_, err := io.ReadAtLeast(r, b, 1)
		return b[0], err
	}
}

// ReadPrefix 读取Reader符合的头部,返回成功(nil),或者错误
func ReadPrefix(r io.Reader, prefix []byte) ([]byte, error) {
	cache := []byte(nil)
	b1 := make([]byte, 1)
	for index := 0; index < len(prefix); {
		switch v := r.(type) {
		case io.ByteReader:
			b, err := v.ReadByte()
			if err != nil {
				return cache, err
			}
			cache = append(cache, b)
		default:
			_, err := io.ReadAtLeast(r, b1, 1)
			if err != nil {
				return cache, err
			}
			cache = append(cache, b1[0])
		}
		if cache[len(cache)-1] == prefix[index] {
			index++
		} else {
			for len(cache) > 0 {
				//only one error in this ReadPrefix ,it is EOF,and not important
				cache2, _ := ReadPrefix(bytes.NewReader(cache[1:]), prefix)
				if len(cache2) > 0 {
					cache = cache2
					break
				}
				cache = cache[1:]
			}
			index = len(cache)
		}
	}
	return cache, nil
}
