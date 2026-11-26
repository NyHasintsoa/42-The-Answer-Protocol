package ui

import (
	"fmt"
	"strings"

	"tap-client/src/util"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"tap-client/src/ui/animation"
	"tap-client/src/ui/theme"
)

// --- View: main screen ------------------------------------------------------------------------------

// mainView renders the protocol log, command input, and session header.
//
// Returns:
//
//	The composed main-screen content.
func (m model) mainView() string {
	p := animation.Painter{Fade: m.phaseFade()}
	cyan, muted := p.C(theme.ColorCyan), p.C(theme.ColorMuted)
	panel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(cyan).
		Padding(0, 1)

	// Log panel: a title row, then rows of [scrollbar][blank][text].
	chip, chipW := m.followChip(p)
	room := m.innerW - chipW - 1
	head := util.TruncateCells("📋 PROTOCOL LOG STREAM", room)
	hint := util.TruncateCells(" (C: Client / S: Server)", room-lipgloss.Width(head))
	left := lipgloss.NewStyle().Bold(true).Foreground(cyan).Render(head) +
		lipgloss.NewStyle().Foreground(muted).Render(hint)
	title := left + strings.Repeat(" ", max(1, m.innerW-lipgloss.Width(left)-chipW)) + chip

	bar := m.scrollbar(p)
	rows := make([]string, 0, m.logH+1)
	rows = append(rows, util.PadTo(title, m.innerW))
	for r := 0; r < m.logH; r++ {
		line := ""
		if idx := m.offset + r; idx < len(m.flat) {
			line = m.flat[idx]
		}
		rows = append(rows, bar[r]+" "+util.PadTo(line, m.textW))
	}
	logPanel := panel.Render(strings.Join(rows, "\n"))

	// Input panel.
	inputHead := "💬 COMMAND INPUT "
	inputHint := util.TruncateCells("(Enter: Send | Wheel/PgUp/PgDn: Scroll | Ctrl+C: Quit)", m.innerW-lipgloss.Width(inputHead))
	inputTitle := lipgloss.NewStyle().Foreground(cyan).Bold(true).Render(inputHead) +
		lipgloss.NewStyle().Foreground(muted).Render(inputHint)
	inputLine := ansi.Truncate(m.input.View(), m.innerW, "")
	inputPanel := panel.Render(util.PadTo(inputTitle, m.innerW) + "\n" + util.PadTo(inputLine, m.innerW))

	return lipgloss.JoinVertical(lipgloss.Left, m.headerView(p), logPanel, inputPanel)
}

// headerView renders the application title, connection badge, and server address.
//
// Parameters:
//
//	p: painter used to apply the current screen fade.
//
// Returns:
//
//	The one-line main-screen header.
func (m model) headerView(p animation.Painter) string {
	pill := lipgloss.NewStyle().Bold(true).
		Foreground(p.C(theme.ColorBg)).Background(p.C(theme.ColorCyan)).Padding(0, 1).
		Render(" THE ANSWER PROTOCOL (RFC 42TAP) ")
	badge := lipgloss.NewStyle().Bold(true).
		Foreground(p.C(theme.ColorBg)).Background(p.C(theme.ColorGreen)).Padding(0, 1).
		Render("CONNECTED")
	server := lipgloss.NewStyle().Foreground(p.C(theme.ColorMuted)).Render("Server: " + m.session.Address())

	var b strings.Builder
	b.WriteString(pill)
	used := lipgloss.Width(pill)
	for _, part := range []string{badge, server} {
		w := lipgloss.Width(part) + 2
		if used+w > m.width {
			break
		}
		b.WriteString("  ")
		b.WriteString(part)
		used += w
	}
	return b.String()
}

// followChip is the status shown at the right of the log title: LIVE while the
// view follows new output, otherwise the scroll position and unread count.
//
// Parameters:
//
//	p: painter used to apply the current screen fade.
//
// Returns:
//
//	The rendered follow-status chip and its terminal-cell width.
func (m model) followChip(p animation.Painter) (string, int) {
	text, c := "● LIVE", animation.RGBGreen
	if !m.follow {
		pct := 100
		if maxOff := m.maxOffset(); maxOff > 0 {
			pct = m.offset * 100 / maxOff
		}
		text, c = fmt.Sprintf("▲ %d%%", pct), animation.RGBYellow
		if m.unread > 0 {
			text += fmt.Sprintf("  ↓ %d new", m.unread)
		}
	}
	chip := lipgloss.NewStyle().Bold(true).Foreground(p.R(c)).Render(text)
	return chip, lipgloss.Width(chip)
}
