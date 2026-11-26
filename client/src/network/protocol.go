package network

import (
	"bufio"
	"time"
)

const dialTimeout = 4 * time.Second

// Connection is the abstraction over a transport used by the terminal UI.
// The UI depends on this interface instead of a concrete *net.Conn, so the
// network implementation can vary without changing the UI (Dependency
// Inversion Principle).
type Connection interface {
	// Send writes a line to the remote peer.
	Send(line string) error
	// Close releases the underlying transport.
	Close() error
	// Reader returns a buffered reader over the transport.
	Reader() *bufio.Reader
}

// ConnectedMsg is delivered when a dial attempt succeeds.
type ConnectedMsg struct {
	Generation int
	Conn       Connection
}

// ServerLineMsg is delivered for each line read from the server.
type ServerLineMsg struct {
	Generation int
	Line       string
}

// ServerErrMsg is delivered when a dial or read operation fails.
type ServerErrMsg struct {
	Generation int
	Err        error
}

// RetryMsg is delivered after the retry delay to trigger a new dial attempt.
type RetryMsg struct{}
