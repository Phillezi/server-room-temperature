package transport

import (
	"machine"
	"machine/usb/cdc"
)

type USB struct {
	dev *cdc.USBCDC
}

func NewUSB() *USB {
	return &USB{
		dev: cdc.New(),
	}
}

func (u *USB) Read(p []byte) (int, error) {
	return u.dev.Read(p)
}

func (u *USB) Write(p []byte) (int, error) {
	return u.dev.Write(p)
}

func (USB) Available() bool {
	return machine.USBCDC.Buffered() > 0
}

func (u *USB) Name() string {
	return "usb"
}
