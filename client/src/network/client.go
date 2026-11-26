package network

import (
	"bufio"
	"context"
	"net"
	"strings"
	"time"
)

// TcpConnection wraps a TCP socket with line buffering.
type TcpConnection struct {
	conn   net.Conn
	reader *bufio.Reader
}

// newTCPConnection wraps conn in a buffered Connection.
//
// Parameters:
//
//	conn: the open TCP socket to wrap.
//
// Returns:
//
//	A transport Connection backed by conn.
func newTCPConnection(conn net.Conn) Connection {
	return &TcpConnection{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}
}

// Send writes line followed by a newline to the peer.
//
// Parameters:
//
//	line: the protocol command to write.
//
// Returns:
//
//	An error if writing the command fails.
func (c *TcpConnection) Send(line string) error {
	_ = c.conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
	_, err := c.conn.Write([]byte(line + "\n"))
	return err
}

// Close closes the underlying TCP socket.
//
// Returns:
//
//	An error if the socket cannot be closed.
func (c *TcpConnection) Close() error {
	return c.conn.Close()
}

// Reader returns the buffered reader for the TCP socket.
//
// Returns:
//
//	The reader associated with this connection.
func (c *TcpConnection) Reader() *bufio.Reader {
	return c.reader
}

// Dial opens a TCP connection to address, subject to ctx and the dial timeout.
//
// Parameters:
//
//	ctx: context used to cancel the connection attempt.
//	address: TCP host and port to connect to.
//
// Returns:
//
//	The connected transport, or an error if dialing fails.
func Dial(ctx context.Context, address string) (Connection, error) {
	dialer := net.Dialer{Timeout: dialTimeout}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, err
	}
	return newTCPConnection(conn), nil
}

// ReadLine reads one line and removes trailing carriage-return and newline characters.
//
// Parameters:
//
//	reader: buffered input from the server connection.
//
// Returns:
//
//	The line, including a final unterminated line when present; otherwise, a read error.
func ReadLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}
