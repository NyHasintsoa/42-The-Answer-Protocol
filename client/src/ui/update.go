package ui

import (
	"bufio"
	"fmt"
	"strings"
	"time"

	"tap-client/src/network"
	"tap-client/src/util"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

// --- Commands ---------------------------------------------------------------------

// Init starts the text-input blink, the first dial attempt and the frame clock.
func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, network.DialCmd(m.address, m.gen), tickCmd())
}

// --- Update -------------------------------------------------------------------------

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
			m.closeConn()
			return m, tea.Quit

		case "enter":
			if !m.connected || m.phase != phaseMain {
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
				m.closeConn()
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

	case network.ConnectedMsg:
		if msg.Generation != m.gen {
			_ = msg.Conn.Close() // a newer attempt superseded this one
			break
		}
		m.conn = msg.Conn
		m.reader = bufio.NewReader(msg.Conn.Reader())
		m.connected = true
		m.lastErr = ""
		m.retryAt = time.Time{}
		m.connectedAt = m.now
		m.revealAt = m.now.Add(successHold + loaderFadeOut)
		m.setPhase(phaseSuccess)
		m.addLog(kindSystem, fmt.Sprintf("Connected to %s", m.address))
		cmds = append(cmds, network.ReadCmd(m.reader, m.gen))

	case network.ServerLineMsg:
		if msg.Generation != m.gen || m.conn == nil {
			break
		}
		m.handleServerLine(msg.Line)
		cmds = append(cmds, network.ReadCmd(m.reader, m.gen))

	case network.ServerErrMsg:
		if msg.Generation != m.gen {
			break
		}
		wasUp := m.connected
		m.closeConn()
		m.connected = false
		m.lastErr = util.DescribeErr(msg.Err)
		m.retryAt = m.now.Add(retryDelay)
		if wasUp {
			m.attempt = 0
			m.connStart = m.now
		}
		if m.phase != phaseLoading {
			m.setPhase(phaseLoading)
		}
		m.addLog(kindError, fmt.Sprintf("Connection error: %v (retrying...)", msg.Err))
		cmds = append(cmds, tea.Tick(retryDelay, func(time.Time) tea.Msg { return network.RetryMsg{} }))

	case network.RetryMsg:
		m.gen++
		m.attempt++
		m.retryAt = time.Time{}
		cmds = append(cmds, network.DialCmd(m.address, m.gen))
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
