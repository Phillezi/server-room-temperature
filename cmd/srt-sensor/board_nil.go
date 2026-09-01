//go:build !rp2040

package main

import "github.com/phillezi/server-room-temperature/pkg/sensor"

type mockSensorImpl struct{}

func (mockSensorImpl) Read() (int32, error) {
	return -1, nil
}

func NewSensor() sensor.Sensor {
	return mockSensorImpl{}
}
