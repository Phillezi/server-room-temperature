package natsconn_test

import (
	"testing"

	"github.com/phillezi/server-room-temperature/internal/natsconn"
)

func TestFormatURL(t *testing.T) {
	tests := []struct {
		name string
		cfg  natsconn.Config
		want string
	}{
		{
			name: "explicit url",
			cfg: natsconn.Config{
				URL: "nats://custom-server:4222",
			},
			want: "nats://custom-server:4222",
		},
		{
			name: "host without scheme",
			cfg: natsconn.Config{
				Host: "remote-nats:4222",
			},
			want: "nats://remote-nats:4222",
		},
		{
			name: "host with scheme",
			cfg: natsconn.Config{
				Host: "tls://remote-nats:4443",
			},
			want: "tls://remote-nats:4443",
		},
		{
			name: "empty config defaults to localhost:4222",
			cfg:  natsconn.Config{},
			want: "nats://localhost:4222",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := natsconn.FormatURL(tt.cfg)
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
