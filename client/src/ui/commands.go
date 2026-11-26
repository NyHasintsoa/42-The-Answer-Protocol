package ui

import (
	appmodel "tap-client/src/model"
	"tap-client/src/service"
	"time"

	tea "charm.land/bubbletea/v2"
)

type tickMsg time.Time

type sessionEventMsg struct {
	event appmodel.ConnectionEvent
}

type sessionClosedMsg struct{}

// sessionEventCmd waits for and converts the next service event into a Bubble Tea message.
//
// Parameters:
//
//	session: service session whose event stream should be observed.
//
// Returns:
//
//	A command that emits the next event message or a closed-session message.
func sessionEventCmd(session *service.Session) tea.Cmd {
	events := session.Start()
	return func() tea.Msg {
		event, ok := <-events
		if !ok {
			return sessionClosedMsg{}
		}
		return sessionEventMsg{event: event}
	}
}

// tickCmd schedules the next UI animation frame.
//
// Returns:
//
//	A command that emits a tickMsg after frameInterval.
func tickCmd() tea.Cmd {
	return tea.Tick(frameInterval, func(t time.Time) tea.Msg { return tickMsg(t) })
}
