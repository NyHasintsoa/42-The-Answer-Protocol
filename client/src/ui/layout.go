package ui

import "charm.land/lipgloss/v2"

// recalculateLayout recomputes the panel geometry from the current terminal
// size and refreshes the log buffer.
//
// The receiver's width and height are the source dimensions; all calculated
// layout values are stored on the receiver.
func (m *model) recalculateLayout() {
	m.innerW = max(8, m.width-4) // panel border (1) + padding (1) on each side
	m.logH = max(1, m.height-chromeRows)
	m.textW = m.innerW - 2 // scrollbar column + one blank column
	m.msgW = max(8, m.textW-prefixW-1)
	m.input.SetWidth(max(4, m.innerW-lipgloss.Width(m.input.Prompt)-1))
	m.rebuild()
}
