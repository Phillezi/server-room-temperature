package main

import (
	"bufio"
	"fmt"
	"io"
)

const maxFrameSize = 32

// FrameReader reads length-delimited frames from a serial stream,
// scanning for the '@' start byte and a '\n' terminator.
type FrameReader struct {
	r *bufio.Reader
}

func NewFrameReader(r io.Reader) *FrameReader {
	return &FrameReader{r: bufio.NewReader(r)}
}

func (fr *FrameReader) ReadFrame(frame []byte) (int, error) {
	if len(frame) < 3 {
		return 0, fmt.Errorf("frame buffer too small")
	}

	n := 0
	for {
		b, err := fr.r.ReadByte()
		if err != nil {
			return 0, err
		}

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
