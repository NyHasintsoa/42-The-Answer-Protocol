package service

import (
	"context"
	"strings"
	"sync"
	"time"

	"tap-client/src/model"
	"tap-client/src/network"
)

const retryDelay = 2 * time.Second

type Session struct {
	address string
	events  chan model.ConnectionEvent

	startOnce  sync.Once
	mu         sync.RWMutex
	cancel     context.CancelFunc
	connection network.Connection
	closed     bool
}

func NewSession(address string) *Session {
	return &Session{
		address: address,
		events:  make(chan model.ConnectionEvent, 16),
	}
}

func (s *Session) Address() string {
	return s.address
}

func (s *Session) Start() <-chan model.ConnectionEvent {
	s.startOnce.Do(func() {
		s.mu.Lock()
		if s.closed {
			s.mu.Unlock()
			close(s.events)
			return
		}
		ctx, cancel := context.WithCancel(context.Background())
		s.cancel = cancel
		s.mu.Unlock()
		go s.run(ctx)
	})
	return s.events
}

func (s *Session) Send(line string) error {
	s.mu.RLock()
	conn := s.connection
	s.mu.RUnlock()
	if conn == nil {
		return network.ErrNotConnected
	}
	return conn.Send(line)
}

func (s *Session) Close() {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.closed = true
	cancel := s.cancel
	conn := s.connection
	s.connection = nil
	s.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	if conn != nil {
		_ = conn.Close()
	}
	if cancel == nil {
		s.startOnce.Do(func() { close(s.events) })
	}
}

func (s *Session) run(ctx context.Context) {
	defer close(s.events)
	defer func() {
		if conn := s.clearConnection(); conn != nil {
			_ = conn.Close()
		}
	}()

	attempt := 1
	for {
		if !s.publish(ctx, model.ConnectionEvent{
			Kind:    model.ConnectionStarting,
			Attempt: attempt,
		}) {
			return
		}

		conn, err := network.Dial(ctx, s.address)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			if !s.publish(ctx, model.ConnectionEvent{
				Kind:    model.ConnectionFailed,
				Attempt: attempt,
				Err:     err,
				RetryAt: time.Now().Add(retryDelay),
			}) || !waitForRetry(ctx) {
				return
			}
			attempt++
			continue
		}

		if !s.setConnection(ctx, conn) {
			_ = conn.Close()
			return
		}
		if !s.publish(ctx, model.ConnectionEvent{
			Kind:    model.ConnectionEstablished,
			Attempt: attempt,
		}) {
			return
		}

		var readErr error
		for {
			line, err := network.ReadLine(conn.Reader())
			if err != nil {
				readErr = err
				break
			}
			if !s.publish(ctx, model.ConnectionEvent{
				Kind: model.ConnectionLine,
				Line: classifyServerLine(line),
			}) {
				return
			}
		}
		if ctx.Err() != nil {
			return
		}

		s.clearConnection()
		_ = conn.Close()
		if !s.publish(ctx, model.ConnectionEvent{
			Kind:         model.ConnectionFailed,
			Err:          readErr,
			RetryAt:      time.Now().Add(retryDelay),
			WasConnected: true,
		}) || !waitForRetry(ctx) {
			return
		}
		attempt = 1
	}
}

func classifyServerLine(text string) model.ServerLine {
	line := model.ServerLine{Kind: model.ServerLineOK, Text: text}
	switch {
	case strings.HasPrefix(text, "EVT "):
		line.Kind = model.ServerLineEvent
	case strings.HasPrefix(text, "ERR "):
		line.Kind = model.ServerLineError
	}
	return line
}

func (s *Session) setConnection(ctx context.Context, conn network.Connection) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || ctx.Err() != nil {
		return false
	}
	s.connection = conn
	return true
}

func (s *Session) clearConnection() network.Connection {
	s.mu.Lock()
	defer s.mu.Unlock()
	conn := s.connection
	s.connection = nil
	return conn
}

func (s *Session) publish(ctx context.Context, event model.ConnectionEvent) bool {
	select {
	case s.events <- event:
		return true
	case <-ctx.Done():
		return false
	}
}

func waitForRetry(ctx context.Context) bool {
	timer := time.NewTimer(retryDelay)
	defer timer.Stop()
	select {
	case <-timer.C:
		return true
	case <-ctx.Done():
		return false
	}
}
