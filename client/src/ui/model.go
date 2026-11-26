package ui

import (
	"fmt"
	"time"

	"tap-client/src/service"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"tap-client/src/ui/theme"
)

// --- Model ------------------------------------------------------------------------

type uiPhase int

const (
	phaseLoading uiPhase = iota // dialling / waiting to retry
	phaseSuccess                // connected: success flash, then the loader dissolves
	phaseMain                   // log + input
)

type model struct {
	session    *service.Session
	connection connectionState

	// Screen state and animation clock.
	width, height int
	now           time.Time
	ticking       bool // a frame tick is already scheduled
	animating     bool // some log line is still animating
	phase         uiPhase
	phaseAt       time.Time
	loaderT0      time.Time

	// Layout, see recalculateLayout.
	innerW int // content width of a panel
	logH   int // visible log rows
	textW  int // width of the log text column (innerW minus scrollbar and gap)
	msgW   int // wrap width of message text

	// Log view.
	entries  []*logEntry
	flat     []string // screen lines of every entry that has started
	lastBorn time.Time
	offset   int // index of the first visible line in flat
	follow   bool
	unread   int
	dragging bool
	dragGrab int // 1/8-cells between the thumb's top and the grabbed point

	input textinput.Model
}

// NewModel builds the initial Bubble Tea model for the server session.
func NewModel(session *service.Session, cfgNote string) tea.Model {
	return initialModel(session, cfgNote)
}

func initialModel(session *service.Session, cfgNote string) model {
	ti := textinput.New()
	ti.Placeholder = "Type command (e.g., CONNECT alice, INVENTORY, USE potion_health_PM, MOVE north)..."
	ti.Focus()
	ti.CharLimit = 512
	ti.SetWidth(80)
	ti.Prompt = "⚡ "

	styles := textinput.DefaultStyles(true)
	styles.Focused.Prompt = lipgloss.NewStyle().Foreground(theme.ColorCyan).Bold(true)
	styles.Blurred.Prompt = styles.Focused.Prompt
	styles.Focused.Text = lipgloss.NewStyle().Foreground(theme.ColorText)
	styles.Blurred.Text = styles.Focused.Text
	ti.SetStyles(styles)

	now := time.Now()
	m := model{
		session: session,
		connection: connectionState{
			attempt:   1,
			connStart: now,
		},
		input:    ti,
		follow:   true,
		ticking:  true, // Init() starts the first frame tick
		phase:    phaseLoading,
		now:      now,
		phaseAt:  now,
		loaderT0: now,
	}
	m.addLog(kindSystem, cfgNote)
	m.addLog(kindSystem, fmt.Sprintf("Connecting to TAP Server at %s...", session.Address()))
	return m
}
