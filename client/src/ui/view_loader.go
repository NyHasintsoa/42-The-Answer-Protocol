package ui

import (
	"fmt"
	"math"
	"strings"

	"tap-client/src/util"

	"charm.land/lipgloss/v2"

	"tap-client/src/ui/animation"
	"tap-client/src/ui/theme"
)

// loaderView renders the animated connection status screen.
//
// Returns:
//
//	The loader card positioned within the terminal.
func (m model) loaderView() string {
	p := animation.Painter{Fade: m.phaseFade()}
	t := math.Max(0, m.now.Sub(m.loaderT0).Seconds())
	age := math.Max(0, m.now.Sub(m.phaseAt).Seconds())
	success := m.phase == phaseSuccess
	failing := !success && m.connection.lastErr != ""
	pulse := (math.Sin(t*2.4) + 1) / 2

	cardW := animation.ClampInt(m.width-12, 30, 46)
	roomy := m.height >= 20
	bigLogo := m.height >= 26
	barW := min(cardW-4, 38)

	ink := func(c animation.RGB) lipgloss.Style { return lipgloss.NewStyle().Foreground(p.R(c)) }

	var lines []string
	add := func(s string) { lines = append(lines, util.Center(s, cardW)) }
	gap := func() {
		if roomy {
			add("")
		}
	}

	title := "THE ANSWER PROTOCOL"
	if cardW >= 40 {
		title = util.SpaceOut(title)
	}
	if bigLogo {
		for row, line := range theme.LogoTAP {
			add(animation.ShimmerLine(line, row, t, p))
		}
		gap()
	}
	add(animation.ShimmerLine(title, len(theme.LogoTAP), t, p))
	add(ink(animation.RGBMuted).Render("RFC 42TAP · TCP client"))
	gap()

	add(loaderBar(barW, t, age, success, p))
	gap()

	spin := theme.SpinnerFrames[int(t*12.5)%len(theme.SpinnerFrames)]
	addr := util.TruncateCells(m.session.Address(), cardW-18)
	var status string
	switch {
	case success:
		status = ink(animation.RGBGreen).Bold(true).Render("✓ Connected to " + addr)
	case failing && m.connection.retryAt.After(m.now):
		wait := m.connection.retryAt.Sub(m.now).Seconds()
		status = ink(animation.RGBYellow).Bold(true).Render(spin) + " " +
			ink(animation.RGBText).Render(fmt.Sprintf("Retrying in %.1fs", wait))
	default:
		dots := int(t*2.5) % 4
		status = ink(animation.RGBYellow).Bold(true).Render(spin) + " " +
			ink(animation.RGBText).Render("Connecting to "+addr) +
			ink(animation.RGBMuted).Render(strings.Repeat(".", dots)+strings.Repeat(" ", 3-dots))
	}
	add(status)

	elapsed := m.now.Sub(m.connection.connStart)
	if success {
		elapsed = m.connection.connectedAt.Sub(m.connection.connStart)
	}
	add(ink(animation.RGBMuted).Render(fmt.Sprintf("attempt %d · %.1fs", m.connection.attempt, elapsed.Seconds())))

	errLine := ""
	if failing {
		errLine = ink(animation.RGBRed).Render(util.TruncateCells("✗ "+m.connection.lastErr, cardW))
	}
	add(errLine)
	gap()

	hint := "Press Ctrl+C to abort"
	if success {
		hint = "Entering session…"
	}
	add(ink(animation.RGBMuted).Italic(true).Render(hint))

	var border animation.RGB
	switch {
	case success:
		border = animation.RGBCyan.Mix(animation.RGBBlue, pulse).Mix(animation.RGBGreen, animation.EaseOutCubic(age/0.25))
	case failing:
		border = animation.RGBYellow.Mix(animation.RGBRed, pulse)
	default:
		border = animation.RGBCyan.Mix(animation.RGBBlue, pulse)
	}
	pad := 0
	if roomy {
		pad = 1
	}
	card := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.R(border)).
		Padding(pad, 3).
		Render(strings.Join(lines, "\n"))
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, card)
}

// loaderBar draws the progress rail: a comet sweeping back and forth while we
// wait, or a green fill once the connection is up.
//
// Parameters:
//
//	width: number of terminal cells occupied by the rail.
//	t: elapsed animation time in seconds.
//	age: elapsed time in the current screen phase, in seconds.
//	success: whether to render the completed connection fill.
//	p: painter used to apply the current screen fade.
//
// Returns:
//
//	The styled progress rail.
func loaderBar(width int, t, age float64, success bool, p animation.Painter) string {
	track := animation.RGBSurface.Mix(animation.RGBBorder, 0.7)
	cell := func(c animation.RGB) string {
		return lipgloss.NewStyle().Foreground(p.R(c)).Render("━")
	}
	var b strings.Builder

	if success {
		fill := animation.EaseOutCubic(age/0.32) * float64(width)
		for i := 0; i < width; i++ {
			b.WriteString(cell(track.Mix(animation.RGBGreen, fill-float64(i))))
		}
		return b.String()
	}

	u := math.Mod(t/1.7, 2)
	pos, dir := u, 1.0
	if u >= 1 {
		pos, dir = 2-u, -1
	}
	head := animation.EaseInOutSine(pos) * float64(width-1)
	tail := 1.2 + 7*math.Sin(math.Pi*pos)
	for i := range width {
		d := (head - float64(i)) * dir
		glow := math.Exp(d * 1.4)
		if d >= 0 {
			glow = math.Exp(-d / tail)
		}
		b.WriteString(cell(track.Mix(animation.RGBCyan, glow).Mix(animation.RGBWhite, math.Pow(glow, 8)*0.75)))
	}
	return b.String()
}
