package component

import (
	"math"
	"unicode/utf8"

	"tap-client/src/util"

	"charm.land/lipgloss/v2"
)

// --- Log buffer & scrolling -----------------------------------------------------------

func (m *model) addLog(kind logKind, text string) {
	if utf8.RuneCountInString(text) > maxLineRunes {
		text = string([]rune(text)[:maxLineRunes]) + "…"
	}

	// Lines logged while the loader is dissolving wait for the log screen, so
	// their entrance animation is actually seen.
	base := m.now
	if m.phase == phaseSuccess && m.revealAt.After(base) {
		base = m.revealAt
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
		m.offset = clampInt(m.offset, 0, m.maxOffset())
	}
}

func (m model) maxOffset() int { return max(0, len(m.flat)-m.logH) }

// scrollTo moves the view; reaching the bottom (re)enables follow mode.
func (m *model) scrollTo(offset int) {
	maxOff := m.maxOffset()
	m.offset = clampInt(offset, 0, maxOff)
	m.follow = m.offset >= maxOff
	if m.follow {
		m.unread = 0
	}
}

func (m *model) scrollBy(delta int) { m.scrollTo(m.offset + delta) }

func (m *model) recalculateLayout() {
	m.innerW = max(8, m.width-4) // panel border (1) + padding (1) on each side
	m.logH = max(1, m.height-chromeRows)
	m.textW = m.innerW - 2 // scrollbar column + one blank column
	m.msgW = max(8, m.textW-prefixW-1)
	m.input.SetWidth(max(4, m.innerW-lipgloss.Width(m.input.Prompt)-1))
	m.rebuild()
}

// --- Mouse: scrollbar click & drag ------------------------------------------------------

func (m *model) mouseDown(x, y int, left bool) {
	if !left || m.phase != phaseMain || m.tooSmall() {
		return
	}
	// Grab area: the scrollbar column plus one column on each side.
	if x < barX-1 || x > barX+1 || y < logTop || y >= logTop+m.logH {
		return
	}
	start, size, ok := thumbSpan(m.logH, len(m.flat), m.offset)
	if !ok {
		return
	}
	pointer := (y-logTop)*8 + 4 // middle of the clicked cell, in 1/8-cell units
	m.dragging = true
	if pointer+4 > start && pointer-4 < start+size { // the clicked cell touches the thumb
		m.dragGrab = pointer - start // exact grab point: nothing moves until the mouse does
		return
	}
	m.dragGrab = size / 2 // clicked the empty rail: centre the thumb on the click
	m.dragTo(y)
}

func (m *model) dragTo(y int) {
	_, size, ok := thumbSpan(m.logH, len(m.flat), m.offset)
	if !ok {
		return
	}
	units := m.logH * 8
	travel := units - size
	if travel <= 0 {
		return
	}
	pointer := clampInt((y-logTop)*8+4, 0, units)
	start := clampInt(pointer-m.dragGrab, 0, travel)
	maxOff := len(m.flat) - m.logH
	m.scrollTo(int(math.Round(float64(start) * float64(maxOff) / float64(travel))))
}
