package protocol_test

import (
	"testing"

	"github.com/phillezi/server-room-temperature/pkg/protocol"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		temp int32
		want string
	}{
		{38124, "@38124\n"},
		{0, "@0\n"},
		{1, "@1\n"},
		{123, "@123\n"},
		{-1234, "@-1234\n"},
	}

	for _, tt := range tests {
		var buf [16]byte

		n := protocol.Encode(buf[:], tt.temp)
		got := string(buf[:n])

		if got != tt.want {
			t.Fatalf(
				"temp=%d: got %q, want %q",
				tt.temp,
				got,
				tt.want,
			)
		}
	}
}
