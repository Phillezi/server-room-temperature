package natsconn

import (
	"fmt"
	"strings"

	"github.com/nats-io/nats.go"
)

type Config struct {
	URL      string `mapstructure:"url"`
	Host     string `mapstructure:"host"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

// FormatURL returns a normalized NATS URL from Config.
func FormatURL(cfg Config) string {
	targetURL := cfg.URL
	if targetURL == "" {
		host := cfg.Host
		if host == "" {
			host = "localhost:4222"
		}
		if !strings.HasPrefix(host, "nats://") &&
			!strings.HasPrefix(host, "tls://") &&
			!strings.HasPrefix(host, "ws://") &&
			!strings.HasPrefix(host, "wss://") {
			targetURL = "nats://" + host
		} else {
			targetURL = host
		}
	}
	return targetURL
}

// Connect creates a NATS connection from the given Config and optional nats.Options.
func Connect(cfg Config, extraOpts ...nats.Option) (*nats.Conn, error) {
	targetURL := FormatURL(cfg)

	var opts []nats.Option
	if cfg.User != "" && cfg.Password != "" {
		opts = append(opts, nats.UserInfo(cfg.User, cfg.Password))
	}
	opts = append(opts, extraOpts...)

	nc, err := nats.Connect(targetURL, opts...)
	if err != nil {
		return nil, fmt.Errorf("connect to nats at %s: %w", targetURL, err)
	}
	return nc, nil
}
