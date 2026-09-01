//go:build rp2040

package main

import (
	"github.com/phillezi/server-room-temperature/pkg/sensor"
)

func NewSensor() sensor.Sensor {
	return sensor.NewInternalTemperature()
}
