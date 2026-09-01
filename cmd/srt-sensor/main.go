package main

import (
	"time"

	"github.com/phillezi/server-room-temperature/pkg/protocol"
	"github.com/phillezi/server-room-temperature/pkg/transport"
)

const samplePeriod = time.Second

func main() {
	s := NewSensor()

	usb := transport.NewUSB()

	next := time.Now()

	for {

		now := time.Now()
		if now.Before(next) {
			time.Sleep(time.Millisecond)
			continue
		}

		next = next.Add(samplePeriod)

		temp, err := s.Read()
		if err != nil {
			continue
		}

		var frame [16]byte
		n := protocol.Encode(frame[:], temp)

		usb.Write(frame[:n])
	}
}
