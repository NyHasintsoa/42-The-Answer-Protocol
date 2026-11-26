package ui

import "tap-client/src/ui/animation"

// needsFrames reports whether anything is still moving. When nothing is, the
// frame clock stops and the program is completely idle.
//
// Returns:
//
//	True while a screen transition or log animation needs another frame.
func (m model) needsFrames() bool {
	if m.phase != phaseMain {
		return true
	}
	return m.now.Sub(m.phaseAt) < mainFadeIn || m.animating
}

// phaseFade is the fade level (0 = invisible, 1 = fully visible) of the screen.
//
// Returns:
//
//	The current phase's opacity in the range [0, 1].
func (m model) phaseFade() float64 {
	age := m.now.Sub(m.phaseAt)
	switch m.phase {
	case phaseLoading:
		return animation.EaseOutCubic(animation.Fraction(age, loaderFadeIn))
	case phaseSuccess:
		if age <= successHold {
			return 1
		}
		return 1 - animation.EaseInCubic(animation.Fraction(age-successHold, loaderFadeOut))
	default:
		return animation.EaseOutCubic(animation.Fraction(age, mainFadeIn))
	}
}

// tooSmall reports whether the terminal is too small to render the UI.
//
// Returns:
//
//	True when either terminal dimension is below its minimum.
func (m model) tooSmall() bool { return m.width < minWidth || m.height < minHeight }
