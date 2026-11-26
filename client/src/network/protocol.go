package network

import (
	"bufio"
	"errors"
	"time"
)

const dialTimeout = 4 * time.Second

var ErrNotConnected = errors.New("not connected")

// Connection abstracts the underlying network transport.
type Connection interface {
	// Send writes a line to the remote peer.
	Send(line string) error
	// Close releases the underlying transport.
	Close() error
	// Reader returns a buffered reader over the transport.
	Reader() *bufio.Reader
}
