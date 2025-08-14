package split

type Checker interface {
	Check([]byte) bool
}

// CRC16Modbus 校验器
type CRC16Modbus struct{}

func (c CRC16Modbus) Check(bs []byte) bool {
	if len(bs) < 3 {
		return false
	}
	payload := bs[:len(bs)-2]
	crc := crc16Modbus(payload)
	crcLow := byte(crc & 0xFF)
	crcHigh := byte(crc >> 8)
	return bs[len(bs)-2] == crcLow && bs[len(bs)-1] == crcHigh
}

// CRC16-Modbus 算法
func crc16Modbus(data []byte) uint16 {
	var crc uint16 = 0xFFFF
	for _, b := range data {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if crc&0x0001 != 0 {
				crc >>= 1
				crc ^= 0xA001
			} else {
				crc >>= 1
			}
		}
	}
	return crc
}

type SumLast struct{}

func (s SumLast) Check(bs []byte) bool {
	if len(bs) < 2 {
		return false
	}
	sum := byte(0)
	for _, b := range bs[:len(bs)-1] {
		sum += b
	}
	return sum == bs[len(bs)-1]
}
