package component

import (
	"fmt"
	"math"
	"strings"

	"tap-client/src/util"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// --- View ------------------------------------------------------------------------------------

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

// --- View: loader --------------------------------------------------------------------------------

func (m model) loaderView() string {
	p := painter{fade: m.phaseFade()}
	t := math.Max(0, m.now.Sub(m.loaderT0).Seconds())
	age := math.Max(0, m.now.Sub(m.phaseAt).Seconds())
	success := m.phase == phaseSuccess
	failing := !success && m.lastErr != ""
	pulse := (math.Sin(t*2.4) + 1) / 2

	cardW := clampInt(m.width-12, 30, 46)
	roomy := m.height >= 20
	bigLogo := m.height >= 26
	barW := min(cardW-4, 38)

	ink := func(c rgb) lipgloss.Style { return lipgloss.NewStyle().Foreground(p.r(c)) }

	var lines []string
	add := func(s string) { lines = append(lines, util.Center(s, cardW)) }
	gap := func() {
		if roomy {
			add("")
		}
	}

	// Title block.
	title := "THE ANSWER PROTOCOL"
	if cardW >= 40 {
		title = util.SpaceOut(title)
	}
	if bigLogo {
		for row, line := range logoTAP {
			add(shimmerLine(line, row, t, p))
		}
		gap()
	}
	add(shimmerLine(title, len(logoTAP), t, p))
	add(ink(rgbMuted).Render("RFC 42TAP · TCP client"))
	gap()

	// Indeterminate scanner, which fills up green once connected.
	add(loaderBar(barW, t, age, success, p))
	gap()

	// Status line.
	spin := spinnerFrames[int(t*12.5)%len(spinnerFrames)]
	addr := util.TruncateCells(m.address, cardW-18)
	var status string
	switch {
	case success:
		status = ink(rgbGreen).Bold(true).Render("✓ Connected to " + addr)
	case failing && m.retryAt.After(m.now):
		wait := m.retryAt.Sub(m.now).Seconds()
		status = ink(rgbYellow).Bold(true).Render(spin) + " " +
			ink(rgbText).Render(fmt.Sprintf("Retrying in %.1fs", wait))
	default:
		dots := int(t*2.5) % 4
		status = ink(rgbYellow).Bold(true).Render(spin) + " " +
			ink(rgbText).Render("Connecting to "+addr) +
			ink(rgbMuted).Render(strings.Repeat(".", dots)+strings.Repeat(" ", 3-dots))
	}
	add(status)

	// Real numbers only: which attempt, and how long we have been at it.
	elapsed := m.now.Sub(m.connStart)
	if success {
		elapsed = m.connectedAt.Sub(m.connStart)
	}
	add(ink(rgbMuted).Render(fmt.Sprintf("attempt %d · %.1fs", m.attempt, elapsed.Seconds())))

	// The last error keeps its line reserved so nothing jumps when it appears.
	errLine := ""
	if failing {
		errLine = ink(rgbRed).Render(util.TruncateCells("✗ "+m.lastErr, cardW))
	}
	add(errLine)
	gap()

	hint := "Press Ctrl+C to abort"
	if success {
		hint = "Entering session…"
	}
	add(ink(rgbMuted).Italic(true).Render(hint))

	// The border pulses cyan/blue, yellow/red while the server is unreachable,
	// and settles on green once connected.
	var border rgb
	switch {
	case success:
		border = rgbCyan.mix(rgbBlue, pulse).mix(rgbGreen, easeOutCubic(age/0.25))
	case failing:
		border = rgbYellow.mix(rgbRed, pulse)
	default:
		border = rgbCyan.mix(rgbBlue, pulse)
	}
	pad := 0
	if roomy {
		pad = 1
	}
	card := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.r(border)).
		Padding(pad, 3).
		Render(strings.Join(lines, "\n"))
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, card)
}

// loaderBar draws the progress rail: a comet sweeping back and forth while we
// wait, or a green fill once the connection is up.
func loaderBar(width int, t, age float64, success bool, p painter) string {
	track := rgbSurface.mix(rgbBorder, 0.7)
	cell := func(c rgb) string {
		return lipgloss.NewStyle().Foreground(p.r(c)).Render("━")
	}
	var b strings.Builder

	if success {
		fill := easeOutCubic(age/0.32) * float64(width)
		for i := 0; i < width; i++ {
			b.WriteString(cell(track.mix(rgbGreen, fill-float64(i))))
		}
		return b.String()
	}

	// Easing slows the comet at both ends and its tail shrinks with its speed,
	// so the turnaround never "jumps".
	u := math.Mod(t/1.7, 2)
	pos, dir := u, 1.0
	if u >= 1 {
		pos, dir = 2-u, -1
	}
	head := easeInOutSine(pos) * float64(width-1)
	tail := 1.2 + 7*math.Sin(math.Pi*pos)
	for i := 0; i < width; i++ {
		d := (head - float64(i)) * dir // > 0: behind the head
		glow := math.Exp(d * 1.4)      // d < 0: short lead-in glow
		if d >= 0 {
			glow = math.Exp(-d / tail)
		}
		b.WriteString(cell(track.mix(rgbCyan, glow).mix(rgbWhite, math.Pow(glow, 8)*0.75)))
	}
	return b.String()
}

// --- View: main screen ------------------------------------------------------------------------------

func (m model) mainView() string {
	p := painter{fade: m.phaseFade()}
	cyan, muted := p.c(colorCyan), p.c(colorMuted)
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

func (m model) headerView(p painter) string {
	pill := lipgloss.NewStyle().Bold(true).
		Foreground(p.c(colorBg)).Background(p.c(colorCyan)).Padding(0, 1).
		Render(" THE ANSWER PROTOCOL (RFC 42TAP) ")
	badge := lipgloss.NewStyle().Bold(true).
		Foreground(p.c(colorBg)).Background(p.c(colorGreen)).Padding(0, 1).
		Render("CONNECTED")
	server := lipgloss.NewStyle().Foreground(p.c(colorMuted)).Render("Server: " + m.address)

	var b strings.Builder
	b.WriteString(pill)
	used := lipgloss.Width(pill)
	for _, part := range []string{badge, server} {
		w := lipgloss.Width(part) + 2
		if used+w > m.width {
			break
		}
		b.WriteString("  ");b.WriteString(part)
		used += w
	}
	return b.String()
}

// followChip is the status shown at the right of the log title: LIVE while the
// view follows new output, otherwise the scroll position and unread count.
func (m model) followChip(p painter) (string, int) {
	text, c := "● LIVE", rgbGreen
	if !m.follow {
		pct := 100
		if maxOff := m.maxOffset(); maxOff > 0 {
			pct = m.offset * 100 / maxOff
		}
		text, c = fmt.Sprintf("▲ %d%%", pct), rgbYellow
		if m.unread > 0 {
			text += fmt.Sprintf("  ↓ %d new", m.unread)
		}
	}
	chip := lipgloss.NewStyle().Bold(true).Foreground(p.r(c)).Render(text)
	return chip, lipgloss.Width(chip)
}

// scrollbar returns one one-cell-wide string per visible log row.
func (m model) scrollbar(p painter) []string {
	start, size, ok := thumbSpan(m.logH, len(m.flat), m.offset)

	trackC := p.r(rgbSurface.mix(rgbBorder, 0.7))
	if !ok {
		trackC = p.r(rgbSurface.mix(rgbBorder, 0.35)) // nothing to scroll: a faint rail
	}
	thumb := rgbCyan
	if m.dragging {
		thumb = rgbCyan.mix(rgbWhite, 0.45)
	}
	thumbC := p.r(thumb)

	cells := make([]string, m.logH)
	for row := range cells {
		glyph, fgThumb, bgThumb := ' ', false, false
		if ok {
			glyph, fgThumb, bgThumb = scrollCell(row, start, size)
		}
		fg, bg := trackC, trackC
		if fgThumb {
			fg = thumbC
		}
		if bgThumb {
			bg = thumbC
		}
		cells[row] = lipgloss.NewStyle().Foreground(fg).Background(bg).Render(string(glyph))
	}
	return cells
}
