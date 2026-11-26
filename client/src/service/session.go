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

// Session owns the server connection lifecycle and publishes events for the UI.
type Session struct {
	address string
	events  chan model.ConnectionEvent

	startOnce  sync.Once
	mu         sync.RWMutex
	cancel     context.CancelFunc
	connection network.Connection
	closed     bool
}

// NewSession creates a session for the server at address.
//
// Parameters:
//
//	address: TCP host and port of the server.
//
// Returns:
//
//	A session ready to start.
func NewSession(address string) *Session {
	return &Session{
		address: address,
		events:  make(chan model.ConnectionEvent, 16),
	}
}

// Address returns the server address configured for the session.
//
// Returns:
//
//	The session's TCP host and port.
func (s *Session) Address() string {
	return s.address
}

// Start begins connection and retry processing once and returns its event stream.
//
// Returns:
//
//	A channel carrying connection-state and server-line events until the session closes.
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

// Send sends a command over the active server connection.
//
// Parameters:
//
//	line: protocol command to send.
//
// Returns:
//
//	An error if the session is disconnected or the write fails.
func (s *Session) Send(line string) error {
	s.mu.RLock()
	conn := s.connection
	s.mu.RUnlock()
	if conn == nil {
		return network.ErrNotConnected
	}
	return conn.Send(line)
}

// Close cancels session processing and closes any active server connection.
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

// run dials, reads server lines, publishes events, and retries until ctx is cancelled.
//
// Parameters:
//
//	ctx: context controlling the lifetime of the session loop.
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

// classifyServerLine assigns a display category based on the protocol prefix.
//
// Parameters:
//
//	text: raw line received from the server.
//
// Returns:
//
//	The line text paired with its event, error, or ordinary-response category.
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

// setConnection stores conn unless the session is closed or ctx is cancelled.
//
// Parameters:
//
//	ctx: context that must remain active for the connection to be accepted.
//	conn: newly established server transport.
//
// Returns:
//
//	True if conn was stored; false if the session is stopping.
func (s *Session) setConnection(ctx context.Context, conn network.Connection) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || ctx.Err() != nil {
		return false
	}
	s.connection = conn
	return true
}

// clearConnection removes and returns the currently stored transport.
//
// Returns:
//
//	The active transport, or nil if no connection is stored.
func (s *Session) clearConnection() network.Connection {
	s.mu.Lock()
	defer s.mu.Unlock()
	conn := s.connection
	s.connection = nil
	return conn
}

// publish sends event unless ctx is cancelled.
//
// Parameters:
//
//	ctx: context used to stop a blocked event send.
//	event: connection or server-line event to publish.
//
// Returns:
//
//	True if event was delivered; false if the context was cancelled first.
func (s *Session) publish(ctx context.Context, event model.ConnectionEvent) bool {
	select {
	case s.events <- event:
		return true
	case <-ctx.Done():
		return false
	}
}

// waitForRetry pauses for retryDelay unless ctx is cancelled.
//
// Parameters:
//
//	ctx: context used to interrupt the retry wait.
//
// Returns:
//
//	True when the delay expires; false when ctx is cancelled.
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
