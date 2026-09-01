package protocol_test

import (
	"testing"

	"github.com/phillezi/server-room-temperature/pkg/protocol"
)

func TestEncodeDecodeRoundtrip(t *testing.T) {
	tests := []int32{
		-2147483648,
		-1234,
		-1,
		0,
		1,
		123,
		38124,
		2147483647,
	}

	for _, want := range tests {
		var buf [16]byte

		n := protocol.Encode(buf[:], want)
		if n == 0 {
			t.Fatalf("temp=%d: Encode failed", want)
		}

		got, ok := protocol.Decode(buf[:n])
		if !ok {
			t.Fatalf("temp=%d: Decode failed for %q", want, buf[:n])
		}

		if got != want {
			t.Fatalf(
				"temp=%d: roundtrip got %d",
				want,
				got,
			)
		}
	}
}
