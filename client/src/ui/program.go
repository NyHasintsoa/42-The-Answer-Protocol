package ui

import tea "charm.land/bubbletea/v2"

// NewProgram builds the Bubbletea program for the TAP terminal client.
func NewProgram(addr, cfgNote string) *tea.Program {
	return tea.NewProgram(NewModel(addr, cfgNote))
}
