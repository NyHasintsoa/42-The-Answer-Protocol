package ui

import (
	"math"

	"tap-client/src/ui/animation"
)

// --- Mouse: scrollbar click & drag ------------------------------------------------------

// mouseDown starts scrollbar dragging or pages the thumb toward the clicked track position.
//
// Parameters:
//
//	x: mouse column in terminal cells.
//	y: mouse row in terminal cells.
//	left: whether the pressed button is the left mouse button.
func (m *model) mouseDown(x, y int, left bool) {
	if !left || m.phase != phaseMain || m.tooSmall() {
		return
	}
	// Grab area: the scrollbar column plus one column on each side.
	if x < barX-1 || x > barX+1 || y < logTop || y >= logTop+m.logH {
		return
	}
	start, size, ok := ThumbSpan(m.logH, len(m.flat), m.offset)
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

// dragTo moves the scrollbar thumb to the supplied terminal row.
//
// Parameters:
//
//	y: mouse row in terminal cells.
func (m *model) dragTo(y int) {
	_, size, ok := ThumbSpan(m.logH, len(m.flat), m.offset)
	if !ok {
		return
	}
	units := m.logH * 8
	travel := units - size
	if travel <= 0 {
		return
	}
	pointer := animation.ClampInt((y-logTop)*8+4, 0, units)
	start := animation.ClampInt(pointer-m.dragGrab, 0, travel)
	maxOff := len(m.flat) - m.logH
	m.scrollTo(int(math.Round(float64(start) * float64(maxOff) / float64(travel))))
}
