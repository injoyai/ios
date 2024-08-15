package ios

import (
	"bytes"
	"io"
	"testing"
)

func TestSplit(t *testing.T) {
	s := &Split{
		Reader: bytes.NewReader([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 1, 2, 3, 5, 6, 7}),
		Check: []Checker{
			&SplitStartEnd{Start: []byte{1, 2, 3}, End: []byte{5, 6, 7}},
			&SplitLength{},
			&SplitTotal{Least: 6},
		},
	}
	for {
		bs, err := s.ReadMessage()
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Error(err)
			return
		}
		t.Log(bs)
	}
}
