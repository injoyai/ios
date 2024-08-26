package ios

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

func NewAReadWithBuffer(buf []byte) func(r Reader) (Acker, error) {
	f := NewReadWithBuffer(buf)
	return func(r Reader) (Acker, error) {
		bs, err := f(r)
		return Ack(bs), err
	}
}

// NewReadWithBuffer 读取函数
func NewReadWithBuffer(buf []byte) func(r Reader) ([]byte, error) {
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
			if buf == nil {
				buf = make([]byte, 1024*4)
			}
			n, err := v.Read(buf)
			if err != nil {
				return nil, err
			}
			return buf[:n], nil

		case io.Reader:
			if buffer == nil {
				buffer = bufio.NewReaderSize(v, 1024*4)
			}
			if buf == nil {
				buf = make([]byte, 1024*4)
			}
			n, err := buffer.Read(buf)
			if err != nil {
				return nil, err
			}
			return buf[:n], nil

		default:
			return nil, fmt.Errorf("未知类型: %T, 未实现[Reader|MReader|AReader]", r)

		}

	}
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
	if i, ok := r.(io.ByteReader); ok {
		return i.ReadByte()
	}
	b := make([]byte, 1)
	_, err := io.ReadAtLeast(r, b, 1)
	return b[0], err
}

// ReadLeast 读取最少least字节,除非返回错误
func ReadLeast(r io.Reader, least int) ([]byte, error) {
	buf := make([]byte, least)
	n, err := io.ReadAtLeast(r, buf, least)
	return buf[:n], err
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
			n, err := r.Read(b1)
			if err != nil {
				return cache, err
			}
			cache = append(cache, b1[:n]...)
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
