// Package service coordinates gameplay operations and communication with the server.
package service

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

type TCPClient struct {
	mu   sync.Mutex
	conn net.Conn
}

// NewTCPClient creates a TCP client with no active connection.
func NewTCPClient() *TCPClient {
	return &TCPClient{}
}

// Connect opens a TCP connection to address and replaces any existing connection.
// ctx controls cancellation; a nil context is treated as context.Background.
// address is the TCP endpoint; a positive timeout limits the dialing operation.
func (client *TCPClient) Connect(ctx context.Context, timeout time.Duration) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	serverAddress := fmt.Sprintf("%s:%s", os.Getenv("SERVER_IP"), os.Getenv("SERVER_PORT"))
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", serverAddress)
	if err != nil {
		return err
	}

	client.mu.Lock()
	if err := ctx.Err(); err != nil {
		client.mu.Unlock()
		_ = conn.Close()
		return err
	}
	previous := client.conn
	client.conn = conn
	client.mu.Unlock()

	if previous != nil {
		_ = previous.Close()
	}
	return nil
}

// Close closes the active connection and clears it from client.
// It is safe to call on a nil client or when no connection is open.
func (client *TCPClient) Close() error {
	if client == nil {
		return nil
	}

	client.mu.Lock()
	conn := client.conn
	client.conn = nil
	client.mu.Unlock()

	if conn == nil {
		return nil
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
