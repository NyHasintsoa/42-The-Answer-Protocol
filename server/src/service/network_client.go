package service

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"unicode/utf8"
)

type NetworkClient struct {
	conn      net.Conn
	writeMu   sync.Mutex
	closeOnce sync.Once
	done      chan struct{}
	Incoming  chan string
	ErrChan   chan error
}

func ConnectTAPServer(ctx context.Context) (*NetworkClient, error) {
	address, err := ServerAddress()
	if err != nil {
		return nil, err
	}
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, fmt.Errorf("connect to TAP server at %s: %w", address, err)
	}
	client := &NetworkClient{
		conn:     conn,
		done:     make(chan struct{}),
		Incoming: make(chan string, 128),
		ErrChan:  make(chan error, 1),
	}
	go client.readLoop()
	return client, nil
}

func (client *NetworkClient) readLoop() {
	defer close(client.Incoming)
	defer close(client.ErrChan)

	reader := bufio.NewReader(client.conn)
	for {
		line, oversized, err := readProtocolLine(reader)
		if err != nil {
			select {
			case <-client.done:
				return
			default:
			}
			select {
			case client.ErrChan <- err:
			default:
			}
			return
		}
		if oversized {
			select {
			case client.ErrChan <- errors.New("server message exceeds the TAP line limit"):
			default:
			}
			return
		}
		line = strings.TrimSuffix(line, "\n")
		line = strings.TrimSuffix(line, "\r")
		select {
		case client.Incoming <- line:
		case <-client.done:
			return
		}
	}
}

func (client *NetworkClient) SendCommand(command string) error {
	if !utf8.ValidString(command) || strings.ContainsAny(command, "\r\n") || len(command)+1 > maxMessageBytes {
		return errors.New("command must be valid UTF-8 and fit on one TAP line")
	}
	client.writeMu.Lock()
	defer client.writeMu.Unlock()
	if client.conn == nil {
		return errors.New("client is closed")
	}
	message := []byte(command + "\n")
	for len(message) > 0 {
		written, err := client.conn.Write(message)
		if err != nil {
			return fmt.Errorf("send TAP command: %w", err)
		}
		if written == 0 {
			return errors.New("send TAP command: connection wrote zero bytes")
		}
		message = message[written:]
	}
	return nil
}

func (client *NetworkClient) Close() {
	client.closeOnce.Do(func() {
		close(client.done)
		if client.conn != nil {
			_ = client.conn.Close()
		}
	})
}
