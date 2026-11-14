package service

import (
	"bufio"
	"net"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
)

const dialTimeout = 4 * time.Second

type ConnectedMsg struct {
	Generation int
	Conn       net.Conn
}

type ServerLineMsg struct {
	Generation int
	Line       string
}

type ServerErrMsg struct {
	Generation int
	Err        error
}

type RetryMsg struct{}

func DialCmd(addr string, generation int) tea.Cmd {
	return func() tea.Msg {
		conn, err := net.DialTimeout("tcp", addr, dialTimeout)
		if err != nil {
			return ServerErrMsg{Generation: generation, Err: err}
		}
		return ConnectedMsg{Generation: generation, Conn: conn}
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

func SendLine(conn net.Conn, line string) error {
	_ = conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
	_, err := conn.Write([]byte(line + "\n"))
	return err
}
