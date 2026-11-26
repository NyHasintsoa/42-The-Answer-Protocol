package model

import "time"

// ConnectionEventKind identifies a session event delivered to the UI.
type ConnectionEventKind uint8

const (
	// ConnectionStarting marks the beginning of a dial attempt.
	ConnectionStarting ConnectionEventKind = iota
	// ConnectionEstablished marks a successful server connection.
	ConnectionEstablished
	// ConnectionLine carries one classified line received from the server.
	ConnectionLine
	// ConnectionFailed marks a dial or active-connection failure.
	ConnectionFailed
)

// ServerLineKind classifies a line received from the server protocol.
type ServerLineKind uint8

const (
	// ServerLineOK is an ordinary server response.
	ServerLineOK ServerLineKind = iota
	// ServerLineEvent is a protocol event line.
	ServerLineEvent
	// ServerLineError is a protocol error line.
	ServerLineError
)

// ServerLine pairs the received text with its protocol category.
type ServerLine struct {
	Kind ServerLineKind
	Text string
}

// ConnectionEvent carries connection status or server data from service to UI.
type ConnectionEvent struct {
	Kind         ConnectionEventKind
	Attempt      int
	Line         ServerLine
	Err          error
	RetryAt      time.Time
	WasConnected bool
}
