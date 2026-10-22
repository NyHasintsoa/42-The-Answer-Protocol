package service

import (
	"context"
	"net"
	"sync"
	"time"
)

type TCPClient struct {
	mu   sync.Mutex
	conn net.Conn
}

func NewTCPClient() *TCPClient {
	return &TCPClient{}
}

func (client *TCPClient) Connect(ctx context.Context, address string, timeout time.Duration) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", address)
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

func (client *TCPClient) IsConnected() bool {
	if client == nil {
		return false
	}
	client.mu.Lock()
	defer client.mu.Unlock()
	return client.conn != nil
}
