package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/phillezi/server-room-temperature/internal/dto"
	"github.com/phillezi/server-room-temperature/pkg/protocol"
)

const (
	maxFrameSize = 32
	natsSubject  = "serverroom.temperature.room1.sensor1"
)

type Publisher interface {
	Publish(subject string, data []byte) error
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s /dev/ttyACM0\n", os.Args[0])
		os.Exit(2)
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "open %s: %v\n", os.Args[1], err)
		os.Exit(1)
	}
	defer f.Close()

	nc, err := nats.Connect(
		"nats://temperature-writer:CHANGE-ME-WRITER@localhost:4222",
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "nats connect: %v\n", err)
		os.Exit(1)
	}
	defer nc.Close()

	var frame [maxFrameSize]byte

	for {
		n, err := readFrame(f, frame[:])
		if err != nil {
			if err == io.EOF {
				return
			}

			fmt.Fprintf(os.Stderr, "read: %v\n", err)
			os.Exit(1)
		}

		milliC, ok := protocol.Decode(frame[:n])
		if !ok {
			continue
		}

		reading := dto.Reading{
			Timestamp:  time.Now().UTC(),
			TempMilliC: milliC,
		}

		data, err := json.Marshal(reading)
		if err != nil {
			fmt.Fprintf(os.Stderr, "marshal reading: %v\n", err)
			continue
		}

		if err := nc.Publish(natsSubject, data); err != nil {
			fmt.Fprintf(os.Stderr, "nats publish: %v\n", err)
			continue
		}

	}
}

func readFrame(r io.Reader, frame []byte) (int, error) {
	if len(frame) < 3 {
		return 0, fmt.Errorf("frame buffer too small")
	}

	var (
		buf [64]byte
		n   int
	)

	for {
		read, err := r.Read(buf[:])

		if read > 0 {
			for _, b := range buf[:read] {
				if n == 0 {
					if b != '@' {
						continue
					}

					frame[0] = '@'
					n = 1
					continue
				}

				if n >= len(frame) {
					if b == '@' {
						frame[0] = '@'
						n = 1
					} else {
						n = 0
					}

					continue
				}

				frame[n] = b
				n++

				if b == '\n' {
					return n, nil
				}
			}
		}

		if err != nil {
			return 0, err
		}
	}
}
