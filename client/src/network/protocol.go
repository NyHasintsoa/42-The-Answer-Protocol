package network

import (
	"bufio"
	"errors"
	"time"
)

const dialTimeout = 4 * time.Second

// ErrNotConnected is returned when a send is attempted without an active connection.
var ErrNotConnected = errors.New("not connected")

// Connection abstracts the underlying network transport.
type Connection interface {
	// Send writes a line to the peer and reports write failures.
	//
	// Parameters:
	//   line: protocol line to send.
	//
	// Returns:
	//   An error if the write fails.
	Send(line string) error
	// Close releases the transport and reports close failures.
	//
	// Returns:
	//   An error if the transport cannot be closed.
	Close() error
	// Reader returns the transport's buffered input reader.
	//
	// Returns:
	//   The buffered reader associated with the connection.
	Reader() *bufio.Reader
}
