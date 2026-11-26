package network

import (
	"bufio"
	"net"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
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

// DialCmd attempts a TCP dial and reports success or failure as a message.
func DialCmd(addr string, generation int) tea.Cmd {
	return func() tea.Msg {
		conn, err := net.DialTimeout("tcp", addr, dialTimeout)
		if err != nil {
			return ServerErrMsg{Generation: generation, Err: err}
		}
		return ConnectedMsg{Generation: generation, Conn: newTCPConnection(conn)}
	}
}

// ReadCmd reads one line; the component re-arms it after each delivered line.
func ReadCmd(reader *bufio.Reader, generation int) tea.Cmd {
	return func() tea.Msg {
		line, err := reader.ReadString('\n')
		if err != nil && line == "" {
			return ServerErrMsg{Generation: generation, Err: err}
		}
		return ServerLineMsg{Generation: generation, Line: strings.TrimRight(line, "\r\n")}
	}
}
