package ui

// phase transitions.

// setPhase changes the active screen phase and records its start time.
//
// Parameters:
//
//	p: phase to activate.
func (m *model) setPhase(p uiPhase) {
	m.phase = p
	m.phaseAt = m.now
}

// advancePhase moves from the success transition to the main screen when its hold and fade finish.
func (m *model) advancePhase() {
	if m.phase == phaseSuccess && m.now.Sub(m.phaseAt) >= successHold+loaderFadeOut {
		m.setPhase(phaseMain)
	}
}
