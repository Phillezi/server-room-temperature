package transport

import "io"

type Transport interface {
	io.ReadWriter
	Name() string
	Available() bool
}
