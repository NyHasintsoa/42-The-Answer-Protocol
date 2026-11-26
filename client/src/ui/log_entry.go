package ui

import (
	"strings"
	"time"
	"unicode/utf8"

	"charm.land/lipgloss/v2"

	"tap-client/src/ui/animation"
	"tap-client/src/util"
)

// logEntry is a single rendered line in the protocol log.
type logEntry struct {
	kind logKind
	text string    // sanitised plain text
	ts   string    // wall-clock arrival time, already formatted
	born time.Time // when its entrance animation starts (may be slightly in the future)

	// Layout cache, rebuilt whenever the width changes.
	wrapW int
	parts []string // text wrapped to wrapW, one element per screen line
	runes int      // total runes in parts (length of the typewriter effect)
	final []string // fully rendered lines once the animation is over
}

func typeDuration(runes int) time.Duration {
	return min(max(time.Duration(runes)*typePerRune, typeMinDur), typeMaxDur)
}

func (e *logEntry) animEnd() time.Duration {
	return max(typeDuration(e.runes), lineFadeDur, labelGlowDur)
}

// render returns the screen lines of the entry at time now, and whether it is
// still animating. The number of lines never changes during the animation.
func (e *logEntry) render(now time.Time, width int) (lines []string, animating bool) {
	if e.parts == nil || e.wrapW != width {
		e.wrapW = width
		e.parts = util.WrapText(e.text, width)
		e.runes = 0
		for _, part := range e.parts {
			e.runes += utf8.RuneCountInString(part)
		}
		e.final = nil
	}
	age := now.Sub(e.born)
	if age >= e.animEnd() {
		if e.final == nil {
			e.final = e.build(e.runes, 1, 1, false)
		}
		return e.final, false
	}
	typing := typeDuration(e.runes)
	shown := int(animation.Clamp01(animation.Fraction(age, typing)) * float64(e.runes))
	fade := animation.EaseOutCubic(animation.Fraction(age, lineFadeDur))
	glow := animation.EaseOutCubic(animation.Fraction(age, labelGlowDur))
	return e.build(shown, fade, glow, age < typing), true
}

// build draws the entry with `shown` runes revealed, text colours faded in by
// fade (0..1) and the label glowing white until glow reaches 1.
func (e *logEntry) build(shown int, fade, glow float64, cursor bool) []string {
	label, labelC, textC := kindStyle(e.kind)
	tsStyle := lipgloss.NewStyle().Foreground(animation.RGBBg.Mix(animation.RGBMuted, fade).ToColor())
	labelStyle := lipgloss.NewStyle().Bold(true).
		Foreground(animation.RGBBg.Mix(labelC, fade).Mix(animation.RGBWhite, (1-glow)*0.9).ToColor())
	textStyle := lipgloss.NewStyle().Foreground(animation.RGBBg.Mix(textC, fade).ToColor())
	indent := strings.Repeat(" ", prefixW)

	out := make([]string, len(e.parts))
	remaining, cursorLine := shown, 0
	for i, part := range e.parts {
		rs := []rune(part)
		take := min(remaining, len(rs))
		remaining -= take
		if take > 0 {
			cursorLine = i
		}
		var b strings.Builder
		if i == 0 {
			b.WriteString(tsStyle.Render(e.ts))
			b.WriteByte(' ')
			b.WriteString(labelStyle.Render(label))
			b.WriteByte(' ')
		} else {
			b.WriteString(indent)
		}
		if take > 0 {
			b.WriteString(textStyle.Render(string(rs[:take])))
		}
		out[i] = b.String()
	}
	if cursor {
		out[cursorLine] += lipgloss.NewStyle().Foreground(animation.RGBCyan.ToColor()).Render("▌")
	}
	return out
}
