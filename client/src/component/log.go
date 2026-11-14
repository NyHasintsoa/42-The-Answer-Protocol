package component

import (
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"tap-client/src/util"

	"charm.land/lipgloss/v2"
)

// --- Log entries --------------------------------------------------------------

type logKind int

const (
	kindSystem logKind = iota
	kindClient
	kindOK
	kindEvent
	kindServerErr
	kindError
)

// kindStyle returns the two-character gutter label of a log kind and the
// colours of that label and of the message text.
func kindStyle(k logKind) (label string, labelC, textC rgb) {
	switch k {
	case kindClient:
		return "C:", rgbBlue, rgbText
	case kindOK:
		return "S:", rgbGreen, rgbGreen
	case kindEvent:
		return "S:", rgbMagenta, rgbMagenta
	case kindServerErr:
		return "S:", rgbRed, rgbRed
	case kindError:
		return "!:", rgbRed, rgbRed
	default:
		return "i:", rgbYellow, rgbYellow
	}
}

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
	shown := int(clamp01(fraction(age, typing)) * float64(e.runes))
	fade := easeOutCubic(fraction(age, lineFadeDur))
	glow := easeOutCubic(fraction(age, labelGlowDur))
	return e.build(shown, fade, glow, age < typing), true
}

// build draws the entry with `shown` runes revealed, text colours faded in by
// fade (0..1) and the label glowing white until glow reaches 1.
func (e *logEntry) build(shown int, fade, glow float64, cursor bool) []string {
	label, labelC, textC := kindStyle(e.kind)
	tsStyle := lipgloss.NewStyle().Foreground(rgbBg.mix(rgbMuted, fade).toColor())
	labelStyle := lipgloss.NewStyle().Bold(true).
		Foreground(rgbBg.mix(labelC, fade).mix(rgbWhite, (1-glow)*0.9).toColor())
	textStyle := lipgloss.NewStyle().Foreground(rgbBg.mix(textC, fade).toColor())
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
		out[cursorLine] += lipgloss.NewStyle().Foreground(rgbCyan.toColor()).Render("▌")
	}
	return out
}

// --- Scrollbar maths ------------------------------------------------------------
//
// The thumb is positioned in 1/8-cell units, so it glides smoothly instead of
// jumping a whole row at a time.

// thumbSpan returns the start and size of the thumb in 1/8-cell units for a
// track of `rows` cells showing `rows` of `total` lines at `offset`. ok is
// false when everything fits and there is nothing to scroll.
func thumbSpan(rows, total, offset int) (start, size int, ok bool) {
	if rows <= 0 || total <= rows {
		return 0, 0, false
	}
	units := rows * 8
	size = clampInt(int(math.Round(float64(units)*float64(rows)/float64(total))), 8, units)
	maxOff := total - rows
	travel := units - size
	start = int(math.Round(float64(travel) * float64(clampInt(offset, 0, maxOff)) / float64(maxOff)))
	return start, size, true
}

var lowerBlocks = [...]rune{' ', '▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// scrollCell describes track row `row` for a thumb at [start, start+size): the
// glyph to draw and whether the thumb colour is used for its foreground and/or
// background (partial cells use a lower-block glyph on two colours).
func scrollCell(row, start, size int) (glyph rune, fgThumb, bgThumb bool) {
	top := row * 8
	covered := min(start+size, top+8) - max(start, top)
	switch {
	case covered <= 0:
		return ' ', false, false // bare track
	case covered >= 8:
		return ' ', false, true // solid thumb
	case start > top: // thumb begins inside this cell and continues below
		return lowerBlocks[covered], true, false
	default: // thumb ends inside this cell
		return lowerBlocks[8-covered], false, true
	}
}
