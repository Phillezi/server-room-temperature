//go:build !tinygo

package transport

import (
	"errors"
)

type USB struct{}

func NewUSB() *USB {
	return &USB{}
}

func (u *USB) Read(p []byte) (int, error) {
	return 0, errors.New("usb cdc transport only supported on tinygo targets")
}

func (u *USB) Write(p []byte) (int, error) {
	return 0, errors.New("usb cdc transport only supported on tinygo targets")
}

func (USB) Available() bool {
	return false
}

func (u *USB) Name() string {
	return "usb"
}
