package model

import "time"

type ConnectionEventKind uint8

const (
	ConnectionStarting ConnectionEventKind = iota
	ConnectionEstablished
	ConnectionLine
	ConnectionFailed
)

type ServerLineKind uint8

const (
	ServerLineOK ServerLineKind = iota
	ServerLineEvent
	ServerLineError
)

type ServerLine struct {
	Kind ServerLineKind
	Text string
}

type ConnectionEvent struct {
	Kind         ConnectionEventKind
	Attempt      int
	Line         ServerLine
	Err          error
	RetryAt      time.Time
	WasConnected bool
}
