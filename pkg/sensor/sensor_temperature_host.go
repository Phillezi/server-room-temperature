//go:build !tinygo

package sensor

import "errors"

type InternalTemperature struct{}

func NewInternalTemperature() *InternalTemperature {
	return &InternalTemperature{}
}

func (InternalTemperature) Read() (int32, error) {
	return 0, errors.New("internal temperature sensor only supported on tinygo targets")
}
