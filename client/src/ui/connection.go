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

func (m *model) sendRaw(line string) error {
	return m.session.Send(line)
}

func (m *model) sendCommand(cmd string) {
	m.addLog(kindClient, cmd)
	if err := m.sendRaw(cmd); err != nil {
		m.addLog(kindError, fmt.Sprintf("Send failed: %v", err))
	}
}

func (m *model) closeSession() {
	m.session.Close()
}

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
