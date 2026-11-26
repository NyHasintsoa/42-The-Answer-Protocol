package network

import (
	"bufio"
	"context"
	"net"
	"strings"
	"time"
)

// tcpConnection wraps a TCP connection and buffers its incoming lines.
type TcpConnection struct {
	conn   net.Conn
	reader *bufio.Reader
}

// newTCPConnection wraps conn into a domain.Connection.
func newTCPConnection(conn net.Conn) Connection {
	return &TcpConnection{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}
}

func (c *TcpConnection) Send(line string) error {
	_ = c.conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
	_, err := c.conn.Write([]byte(line + "\n"))
	return err
}

func (c *TcpConnection) Close() error {
	return c.conn.Close()
}

func (c *TcpConnection) Reader() *bufio.Reader {
	return c.reader
}

func Dial(ctx context.Context, address string) (Connection, error) {
	dialer := net.Dialer{Timeout: dialTimeout}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, err
	}
	return newTCPConnection(conn), nil
}

func ReadLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}
