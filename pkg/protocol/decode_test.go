package protocol_test

import (
	"testing"

	"github.com/phillezi/server-room-temperature/pkg/protocol"
)

func TestDecode(t *testing.T) {
	tests := []struct {
		input string
		want  int32
		ok    bool
	}{
		{"@38124\n", 38124, true},
		{"@0\n", 0, true},
		{"@1\n", 1, true},
		{"@123\n", 123, true},
		{"@-1234\n", -1234, true},
		{"@2147483647\n", 2147483647, true},
		{"@-2147483648\n", -2147483648, true},

		{"", 0, false},
		{"@\n", 0, false},
		{"@-\n", 0, false},
		{"@12", 0, false},
		{"12\n", 0, false},
		{"@12x\n", 0, false},
		{"@2147483648\n", 0, false},
		{"@-2147483649\n", 0, false},
	}

	for _, tt := range tests {
		got, ok := protocol.Decode([]byte(tt.input))

		if ok != tt.ok || got != tt.want {
			t.Fatalf(
				"input=%q: got (%d, %t), want (%d, %t)",
				tt.input,
				got,
				ok,
				tt.want,
				tt.ok,
			)
		}
	}
}
