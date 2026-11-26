package ui

import (
	"fmt"
	"strings"
	"time"

	appmodel "tap-client/src/model"
	"tap-client/src/util"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

// --- Commands ---------------------------------------------------------------------

// Init starts the text-input blink, the first dial attempt and the frame clock.
//
// Returns:
//
//	Batched commands for input blinking, service events, and animation frames.
func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, sessionEventCmd(m.session), tickCmd())
}

// --- Update -------------------------------------------------------------------------

// Update processes a Bubble Tea message and returns the updated model and follow-up command.
//
// Parameters:
//
//	msg: keyboard, mouse, window, timer, or service-event message to process.
//
// Returns:
//
//	The updated UI model and any command that should run next.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	m.now = time.Now()

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if m.width == 0 {
			m.phaseAt = m.now // start the loader fade-in once we can actually draw it
		}
		m.width, m.height = msg.Width, msg.Height
		m.recalculateLayout()

	case tickMsg:
		m.ticking = false // this tick has fired; a new one is scheduled below if needed
		m.advancePhase()
		m.rebuild()

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			_ = m.sendRaw("QUIT")
			m.closeSession()
			return m, tea.Quit

		case "enter":
			if !m.connection.connected || m.phase != phaseMain {
				return m, nil
			}
			raw := strings.TrimSpace(m.input.Value())
			if raw == "" {
				return m, nil
			}
			m.input.SetValue("")
			m.sendCommand(raw)
			m.scrollTo(m.maxOffset()) // jump to the newest output
			if strings.EqualFold(raw, "QUIT") {
				m.closeSession()
				return m, tea.Quit
			}

		case "pgup":
			m.scrollBy(-max(1, m.logH-1))
		case "pgdown":
			m.scrollBy(max(1, m.logH-1))
		case "ctrl+home":
			m.scrollTo(0)
		case "ctrl+end":
			m.scrollTo(m.maxOffset())
		}

	case tea.MouseWheelMsg:
		if m.phase == phaseMain {
			switch msg.Button {
			case tea.MouseWheelUp:
				m.scrollBy(-wheelStep)
			case tea.MouseWheelDown:
				m.scrollBy(wheelStep)
			}
		}

	case tea.MouseClickMsg:
		m.mouseDown(msg.X, msg.Y, msg.Button == tea.MouseLeft)

	case tea.MouseMotionMsg:
		if m.dragging {
			m.dragTo(msg.Y)
		}

	case tea.MouseReleaseMsg:
		m.dragging = false

	case sessionEventMsg:
		event := msg.event
		switch event.Kind {
		case appmodel.ConnectionStarting:
			m.connection.attempt = event.Attempt
			m.connection.retryAt = time.Time{}

		case appmodel.ConnectionEstablished:
			m.connection.connected = true
			m.connection.lastErr = ""
			m.connection.retryAt = time.Time{}
			m.connection.connectedAt = m.now
			m.connection.revealAt = m.now.Add(successHold + loaderFadeOut)
			m.setPhase(phaseSuccess)
			m.addLog(kindSystem, fmt.Sprintf("Connected to %s", m.session.Address()))

		case appmodel.ConnectionLine:
			if m.connection.connected {
				m.handleServerLine(event.Line)
			}

		case appmodel.ConnectionFailed:
			wasUp := event.WasConnected || m.connection.connected
			m.connection.connected = false
			m.connection.lastErr = util.DescribeErr(event.Err)
			m.connection.retryAt = event.RetryAt
			if wasUp {
				m.connection.attempt = 0
				m.connection.connStart = m.now
			}
			if m.phase != phaseLoading {
				m.setPhase(phaseLoading)
			}
			m.addLog(kindError, fmt.Sprintf("Connection error: %v (retrying...)", event.Err))
		}
		cmds = append(cmds, sessionEventCmd(m.session))

	case sessionClosedMsg:
		return m, tea.Quit
	}

	var inputCmd tea.Cmd
	m.input, inputCmd = m.input.Update(msg)
	cmds = append(cmds, inputCmd)

	if m.needsFrames() && !m.ticking {
		m.ticking = true
		cmds = append(cmds, tickCmd())
	}
	return m, tea.Batch(cmds...)
}
