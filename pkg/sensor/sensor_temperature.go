package sensor

import "machine"

type InternalTemperature struct{}

func NewInternalTemperature() *InternalTemperature {
	return &InternalTemperature{}
}

func (InternalTemperature) Read() (int32, error) {
	return machine.ReadTemperature(), nil
}
