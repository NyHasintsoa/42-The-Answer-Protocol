package ui

// phase transitions.

func (m *model) setPhase(p uiPhase) {
	m.phase = p
	m.phaseAt = m.now
}

func (m *model) advancePhase() {
	if m.phase == phaseSuccess && m.now.Sub(m.phaseAt) >= successHold+loaderFadeOut {
		m.setPhase(phaseMain)
	}
}
