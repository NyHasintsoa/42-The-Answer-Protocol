package ui

import (
	"math"
	"unicode/utf8"

	"tap-client/src/ui/animation"
	"tap-client/src/util"
)

// --- Log buffer & scrolling -----------------------------------------------------------

// addLog appends a new entry to the log and schedules a rebuild. Lines logged
// while the loader is dissolving wait for the log screen, so their entrance
// animation is actually seen.
func (m *model) addLog(kind logKind, text string) {
	if utf8.RuneCountInString(text) > maxLineRunes {
		text = string([]rune(text)[:maxLineRunes]) + "…"
	}

	base := m.now
	if m.phase == phaseSuccess && m.connection.revealAt.After(base) {
		base = m.connection.revealAt
	}
	born := base
	if next := m.lastBorn.Add(lineStagger); next.After(born) {
		born = next
	}
	if limit := base.Add(maxStaggerLead); born.After(limit) {
		born = limit
	}
	m.lastBorn = born

	m.entries = append(m.entries, &logEntry{
		kind: kind,
		text: util.Sanitize(text),
		ts:   m.now.Format(stampLayout),
		born: born,
	})
	if !m.follow {
		m.unread++
	}
	m.trimEntries()
	m.rebuild()
}

// trimEntries drops the oldest entries beyond maxLogEntries. When the user has
// scrolled up, the offset is shifted so the lines they are reading stay put.
func (m *model) trimEntries() {
	extra := len(m.entries) - maxLogEntries
	if extra <= 0 {
		return
	}
	removed := 0
	for _, e := range m.entries[:extra] {
		removed += max(1, len(e.parts))
	}
	m.entries = append([]*logEntry(nil), m.entries[extra:]...)
	if !m.follow {
		m.offset = max(0, m.offset-removed)
	}
}

// rebuild re-renders every entry that has started into m.flat. It is cheap:
// finished entries come from their cache and only animating ones are redrawn.
func (m *model) rebuild() {
	if m.msgW <= 0 {
		return // no size yet
	}
	flat := make([]string, 0, len(m.flat)+4)
	animating := false
	for _, e := range m.entries {
		if e.born.After(m.now) { // waiting for its turn (stagger / log screen reveal)
			animating = true
			break
		}
		lines, moving := e.render(m.now, m.msgW)
		animating = animating || moving
		flat = append(flat, lines...)
	}
	m.flat = flat
	m.animating = animating
	if m.follow {
		m.offset = m.maxOffset()
	} else {
		m.offset = animation.ClampInt(m.offset, 0, m.maxOffset())
	}
}

func (m model) maxOffset() int { return max(0, len(m.flat)-m.logH) }

// scrollTo moves the view; reaching the bottom (re)enables follow mode.
func (m *model) scrollTo(offset int) {
	maxOff := m.maxOffset()
	m.offset = animation.ClampInt(offset, 0, maxOff)
	m.follow = m.offset >= maxOff
	if m.follow {
		m.unread = 0
	}
}

func (m *model) scrollBy(delta int) { m.scrollTo(m.offset + delta) }

// --- Scrollbar geometry ------------------------------------------------------------

// lowerBlocks maps a covered fraction of a cell to a lower-block glyph.
var lowerBlocks = [...]rune{' ', '▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// ThumbSpan returns the start and size of the thumb in 1/8-cell units for a
// track of `rows` cells showing `rows` of `total` lines at `offset`. ok is
// false when everything fits and there is nothing to scroll.
func ThumbSpan(rows, total, offset int) (start, size int, ok bool) {
	if rows <= 0 || total <= rows {
		return 0, 0, false
	}
	units := rows * 8
	size = animation.ClampInt(int(math.Round(float64(units)*float64(rows)/float64(total))), 8, units)
	maxOff := total - rows
	travel := units - size
	start = int(math.Round(float64(travel) * float64(animation.ClampInt(offset, 0, maxOff)) / float64(maxOff)))
	return start, size, true
}

// ScrollCell describes track row `row` for a thumb at [start, start+size): the
// glyph to draw and whether the thumb colour is used for its foreground and/or
// background (partial cells use a lower-block glyph on two colours).
func ScrollCell(row, start, size int) (glyph rune, fgThumb, bgThumb bool) {
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
