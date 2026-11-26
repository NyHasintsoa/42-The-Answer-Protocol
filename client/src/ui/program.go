package ui

import (
	"tap-client/src/service"

	tea "charm.land/bubbletea/v2"
)

// NewProgram builds the Bubbletea program for the TAP terminal client.
//
// Parameters:
//
//	session: service session used by the UI model.
//	cfgNote: startup note displayed in the protocol log.
//
// Returns:
//
//	A Bubble Tea program configured with the TAP UI model.
func NewProgram(session *service.Session, cfgNote string) *tea.Program {
	return tea.NewProgram(NewModel(session, cfgNote))
}
