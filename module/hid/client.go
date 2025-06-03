package hid

import (
	"context"
	"fmt"
	"github.com/injoyai/ios"
	"github.com/karalabe/hid"
)

func NewDial(vendorID, productID uint16) ios.DialFunc {
	return func(ctx context.Context) (ios.ReadWriteCloser, string, error) {
		device, err := Dial(vendorID, productID)
		return device, fmt.Sprintf("%04x-%4x", vendorID, productID), err
	}
}

func Dial(vendorID, productID uint16) (*hid.Device, error) {
	devs := hid.Enumerate(vendorID, productID)
	if len(devs) == 0 {
		return nil, fmt.Errorf("未找到目标HID设备")
	}
	return devs[0].Open()
}

func List() []hid.DeviceInfo {
	// 0,0 表示列出所有设备
	return hid.Enumerate(0, 0)
}
