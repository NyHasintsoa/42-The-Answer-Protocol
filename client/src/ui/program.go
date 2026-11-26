package ui

import (
	"tap-client/src/service"

	tea "charm.land/bubbletea/v2"
)

// NewProgram builds the Bubbletea program for the TAP terminal client.
func NewProgram(session *service.Session, cfgNote string) *tea.Program {
	return tea.NewProgram(NewModel(session, cfgNote))
}
