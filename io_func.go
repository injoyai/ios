package ios

import (
	"io"
)

// SplitBytesByLength
// 按最大长度分割字节 todo 这个不应该出现在这里
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

func CheckReader(r Reader) error {
	switch r.(type) {
	case AReader, MReader, io.Reader:
		return nil
	default:
		return ErrUnknownReader
	}
}
