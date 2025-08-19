package split

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"testing"
)

func TestSplit_ReadFrom(t *testing.T) {
	{
		fn := &Split{
			Prefixes: []Prefix{{0x03}},
			Suffix:   nil,
			Length: Length{
				LittleEndian: false,
				Start:        1,
				End:          1,
				Fixed:        3,
			},
			Checker: nil,
			OnErr:   nil,
		}
		buf := bufio.NewReader(bytes.NewBuffer([]byte{0x01, 0x03, 0x03, 0x11, 0x011, 0x04, 0x04, 0x05}))
		val, err := fn.ReadFrom(buf)
		if err != nil {
			t.Error(err)
		}
		if hex.EncodeToString(val) != hex.EncodeToString([]byte{0x03, 0x03, 0x11, 0x011, 0x04, 0x04}) {
			t.Error("测试失败: " + hex.EncodeToString(val))
		} else {
			t.Log("测试通过")
		}
	}
}
