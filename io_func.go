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

func A2MReader(r AReader) MReader {
	return MReadFunc(func() ([]byte, error) {
		a, err := r.ReadAck()
		if err != nil {
			return nil, err
		}
		return a.Payload(), nil
	})
}

func M2AReader(r MReader) AReader {
	return AReadFunc(func() (Acker, error) {
		m, err := r.ReadMessage()
		if err != nil {
			return nil, err
		}
		return Ack(m), nil
	})
}

func NewIO(r io.Reader, w io.Writer, c io.Closer) IO {
	return struct {
		io.Reader
		io.Writer
		io.Closer
	}{
		Reader: r,
		Writer: w,
		Closer: c,
	}
}

func NewMIO(r MReader, w io.Writer, c io.Closer) MIO {
	return struct {
		MReader
		io.Writer
		io.Closer
	}{
		MReader: r,
		Writer:  w,
		Closer:  c,
	}
}

func NewAIO(r AReader, w io.Writer, c io.Closer) AIO {
	return struct {
		AReader
		io.Writer
		io.Closer
	}{
		AReader: r,
		Writer:  w,
		Closer:  c,
	}
}
