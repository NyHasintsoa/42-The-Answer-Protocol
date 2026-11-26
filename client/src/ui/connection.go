package ui

import (
	"fmt"
	"time"

	appmodel "tap-client/src/model"
)

// --- Connection -------------------------------------------------------------------

type connectionState struct {
	attempt     int
	connected   bool
	lastErr     string
	retryAt     time.Time
	connStart   time.Time
	connectedAt time.Time
	revealAt    time.Time
}

// sendRaw sends line through the service session without changing the log.
//
// Parameters:
//
//	line: raw protocol command to send.
//
// Returns:
//
//	An error if the session is disconnected or the write fails.
func (m *model) sendRaw(line string) error {
	return m.session.Send(line)
}

// sendCommand logs cmd as client input and sends it to the server.
//
// Parameters:
//
//	cmd: protocol command entered by the user.
func (m *model) sendCommand(cmd string) {
	m.addLog(kindClient, cmd)
	if err := m.sendRaw(cmd); err != nil {
		m.addLog(kindError, fmt.Sprintf("Send failed: %v", err))
	}
}

// closeSession stops the service session and releases its active connection.
func (m *model) closeSession() {
	m.session.Close()
}

// handleServerLine adds a classified server line to the protocol log.
//
// Parameters:
//
//	line: server text paired with its protocol category.
func (m *model) handleServerLine(line appmodel.ServerLine) {
	switch line.Kind {
	case appmodel.ServerLineEvent:
		m.addLog(kindEvent, line.Text)
	case appmodel.ServerLineError:
		m.addLog(kindServerErr, line.Text)
	case appmodel.ServerLineOK:
		m.addLog(kindOK, line.Text)
	default:
		m.addLog(kindError, fmt.Sprintf("Unknown server line kind %d: %s", line.Kind, line.Text))
	}
}
