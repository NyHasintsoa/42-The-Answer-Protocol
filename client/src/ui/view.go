package ui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// --- View ------------------------------------------------------------------------------------

// View renders the active screen or a message explaining why it cannot be rendered.
//
// Returns:
//
//	The Bubble Tea view with alternate-screen and mouse interaction enabled.
func (m model) View() tea.View {
	var content string
	switch {
	case m.width == 0:
		content = "Initializing TAP Terminal UI..."
	case m.tooSmall():
		content = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			fmt.Sprintf("Terminal too small (%dx%d)\nneeds at least %dx%d", m.width, m.height, minWidth, minHeight))
	case m.phase == phaseMain:
		content = m.mainView()
	default:
		content = m.loaderView()
	}

	v := tea.NewView(content)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion // wheel + click/drag on the scrollbar
	return v
}
