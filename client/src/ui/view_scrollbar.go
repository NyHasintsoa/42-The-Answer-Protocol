package ui

import (
	"charm.land/lipgloss/v2"

	"tap-client/src/ui/animation"
)

// --- View: scrollbar -----------------------------------------------------------------------------

// scrollbar returns one one-cell-wide string per visible log row.
func (m model) scrollbar(p animation.Painter) []string {
	start, size, ok := ThumbSpan(m.logH, len(m.flat), m.offset)

	trackC := p.R(animation.RGBSurface.Mix(animation.RGBBorder, 0.7))
	if !ok {
		trackC = p.R(animation.RGBSurface.Mix(animation.RGBBorder, 0.35)) // nothing to scroll: a faint rail
	}
	thumb := animation.RGBCyan
	if m.dragging {
		thumb = thumb.Mix(animation.RGBWhite, 0.45)
	}
	thumbC := p.R(thumb)

	cells := make([]string, m.logH)
	for row := range cells {
		glyph, fgThumb, bgThumb := ' ', false, false
		if ok {
			glyph, fgThumb, bgThumb = ScrollCell(row, start, size)
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
