package common

import "fmt"

type Address interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~string
}

func ToAddress[T Address](addr T) string {
	address := ""
	switch v := any(addr).(type) {
	case string:
		address = v
	default:
		address = fmt.Sprintf(":%d", v)
	}
	return address
}
