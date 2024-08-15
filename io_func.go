package ios

import (
	"bytes"
	"io"
)

// ReadLeast 读取最少least字节,除非返回错误
func ReadLeast(r io.Reader, least int) ([]byte, error) {
	buf := make([]byte, least)
	n, err := io.ReadAtLeast(r, buf, least)
	return buf[:n], err
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

// ReadPrefix 读取Reader符合的头部,返回成功(nil),或者错误
func ReadPrefix(r io.Reader, prefix []byte) ([]byte, error) {
	cache := []byte(nil)
	for index := 0; index < len(prefix); {
		b, err := ReadByte(r)
		if err != nil {
			return cache, err
		}
		cache = append(cache, b)
		if b == prefix[index] {
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

// SplitBytesByLength 按最大长度分割字节
func SplitBytesByLength(p []byte, max int) [][]byte {
	if max == 0 {
		return [][]byte{}
	}
	list := [][]byte(nil)
	for len(p) > max {
		list = append(list, p[:max])
		p = p[max:]
	}
	list = append(list, p)
	return list
}
