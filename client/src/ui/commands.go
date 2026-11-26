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

func tickCmd() tea.Cmd {
	return tea.Tick(frameInterval, func(t time.Time) tea.Msg { return tickMsg(t) })
}
