// Package service coordinates gameplay operations and communication with the server.
package service

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

type TCPClient struct {
	mu        sync.Mutex
	request   sync.Mutex
	conn      net.Conn
	events    chan string
	responses chan string
	done      chan struct{}
	stop      func()
}

// NewTCPClient creates a TCP client with no active connection.
func NewTCPClient() *TCPClient {
	return &TCPClient{events: make(chan string, 64)}
}

// Connect opens a TCP connection using SERVER_IP and SERVER_PORT and replaces any existing connection.
// ctx controls cancellation; a nil context is treated as context.Background. A positive timeout
// limits dialing and reading the server greeting.
func (client *TCPClient) Connect(ctx context.Context, timeout time.Duration) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	serverAddress, err := ServerAddress()
	if err != nil {
		return err
	}
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", serverAddress)
	if err != nil {
		return err
	}
	if err := conn.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
		_ = conn.Close()
		return err
	}
	reader := bufio.NewReader(conn)
	greeting, err := reader.ReadString('\n')
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("read server greeting: %w", err)
	}
	if strings.TrimSpace(greeting) != "OK hello proto=1" {
		_ = conn.Close()
		return fmt.Errorf("unexpected server greeting %q", strings.TrimSpace(greeting))
	}
	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		_ = conn.Close()
		return err
	}

	client.mu.Lock()
	if err := ctx.Err(); err != nil {
		client.mu.Unlock()
		_ = conn.Close()
		return err
	}
	previous := client.conn
	previousStop := client.stop
	stop := make(chan struct{})
	var stopOnce sync.Once
	stopConnection := func() { stopOnce.Do(func() { close(stop) }) }
	client.conn = conn
	client.events = make(chan string, 64)
	client.responses = make(chan string, 1)
	client.done = stop
	client.stop = stopConnection
	events := client.events
	responses := client.responses
	client.mu.Unlock()

	if previous != nil {
		if previousStop != nil {
			previousStop()
		}
		_ = previous.Close()
	}
	go client.readLoop(conn, reader, events, responses, stop, stopConnection)
	return nil
}

// Command sends one TAP command and returns its response. EVT messages received
// while waiting are exposed through Events instead of being mistaken for replies.
func (client *TCPClient) Command(ctx context.Context, command string) (string, error) {
	command = strings.TrimSpace(command)
	if command == "" || strings.ContainsAny(command, "\r\n") {
		return "", errors.New("command must contain one non-empty protocol line")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	client.request.Lock()
	defer client.request.Unlock()

	client.mu.Lock()
	conn, responses, done := client.conn, client.responses, client.done
	client.mu.Unlock()
	if conn == nil || responses == nil || done == nil {
		return "", errors.New("not connected to the server")
	}
	if err := conn.SetWriteDeadline(time.Now().Add(3 * time.Second)); err != nil {
		return "", err
	}
	defer conn.SetWriteDeadline(time.Time{})

	message := []byte(command + "\n")
	for len(message) > 0 {
		written, err := conn.Write(message)
		if err != nil {
			_ = client.Close()
			return "", fmt.Errorf("send command: %w", err)
		}
		if written == 0 {
			_ = client.Close()
			return "", errors.New("server connection wrote zero bytes")
		}
		message = message[written:]
	}

	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	select {
	case response := <-responses:
		return response, nil
	case <-ctx.Done():
		_ = client.Close()
		return "", ctx.Err()
	case <-timer.C:
		_ = client.Close()
		return "", errors.New("timed out waiting for server response")
	case <-done:
		select {
		case response := <-responses:
			return response, nil
		default:
			return "", errors.New("server connection closed while waiting for response")
		}
	}
}

func (client *TCPClient) readLoop(conn net.Conn, reader *bufio.Reader, events, responses chan string, done <-chan struct{}, stop func()) {
	defer func() {
		stop()
		_ = conn.Close()
		client.mu.Lock()
		if client.conn == conn {
			client.conn = nil
			client.stop = nil
		}
		client.mu.Unlock()
	}()

	for {
		line, err := reader.ReadString('\n')
		if line != "" {
			line = strings.TrimRight(line, "\r\n")
			target := responses
			if strings.HasPrefix(line, "EVT ") {
				target = events
			}
			select {
			case target <- line:
			case <-done:
				return
			}
		}
		if err != nil {
			return
		}
	}
}

// Events returns server-pushed EVT messages received while awaiting responses.
func (client *TCPClient) Events() <-chan string {
	client.mu.Lock()
	defer client.mu.Unlock()
	return client.events
}

// Close closes the active connection and clears it from client.
// It is safe to call on a nil client or when no connection is open.
func (client *TCPClient) Close() error {
	if client == nil {
		return nil
	}

	client.mu.Lock()
	conn := client.conn
	stop := client.stop
	client.conn = nil
	client.stop = nil
	client.mu.Unlock()

	if conn == nil {
		return nil
	}
	if stop != nil {
		stop()
	}
	return conn.Close()
}

// IsConnected reports whether client currently holds a connection.
func (client *TCPClient) IsConnected() bool {
	if client == nil {
		return false
	}
	client.mu.Lock()
	defer client.mu.Unlock()
	return client.conn != nil
}
