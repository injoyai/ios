package split

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestSplit_ReadFrom(t *testing.T) {
	{
		fn := &Split{
			Checker: []Checker{
				Prefix{0x03},
				Length{
					Start: 1,
					End:   1,
					Fixed: 3,
				},
			},
		}
		buf := bytes.NewBuffer([]byte{0x01, 0x03, 0x03, 0x11, 0x011, 0x04, 0x04, 0x05})
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

func TestLength_Check(t *testing.T) {
	c := Length{
		Start: 1,
		End:   1,
		Fixed: 3,
	}
	match, invalid, err := c.Check([]byte{0x03, 0x03, 0x11, 0x011, 0x04, 0x04})
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(match, invalid)
}
