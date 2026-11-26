package ui

import (
	"errors"
	"fmt"
	"strings"
)

// --- Connection -------------------------------------------------------------------

func (m *model) sendRaw(line string) error {
	if m.conn == nil {
		return errors.New("not connected")
	}
	return m.conn.Send(line)
}

func (m *model) sendCommand(cmd string) {
	m.addLog(kindClient, cmd)
	if err := m.sendRaw(cmd); err != nil {
		m.addLog(kindError, fmt.Sprintf("Send failed: %v", err))
	}
}

func (m *model) closeConn() {
	if m.conn != nil {
		_ = m.conn.Close()
		m.conn = nil
	}
	m.reader = nil
}

func (m *model) handleServerLine(line string) {
	switch {
	case strings.HasPrefix(line, "EVT "):
		m.addLog(kindEvent, line)
	case strings.HasPrefix(line, "ERR "):
		m.addLog(kindServerErr, line)
	default:
		m.addLog(kindOK, line)
	}
}
