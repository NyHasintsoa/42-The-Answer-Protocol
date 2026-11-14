package component

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// --- Model ------------------------------------------------------------------------

type uiPhase int

const (
	phaseLoading uiPhase = iota // dialling / waiting to retry
	phaseSuccess                // connected: success flash, then the loader dissolves
	phaseMain                   // log + input
)

type model struct {
	// Connection.
	address     string
	conn        net.Conn
	reader      *bufio.Reader
	gen         int // id of the current dial attempt
	attempt     int
	connected   bool
	lastErr     string
	retryAt     time.Time // zero unless we are waiting to retry
	connStart   time.Time // start of the current connecting sequence
	connectedAt time.Time
	revealAt    time.Time // when the log screen appears after a successful connect

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

func NewModel(addr, cfgNote string) tea.Model {
	return initialModel(addr, cfgNote)
}

func initialModel(addr, cfgNote string) model {
	ti := textinput.New()
	ti.Placeholder = "Type command (e.g., CONNECT alice, LOOK, MOVE north, CHAT GLOBAL hello)..."
	ti.Focus()
	ti.CharLimit = 512
	ti.SetWidth(80)
	ti.Prompt = "⚡ "

	styles := textinput.DefaultStyles(true)
	styles.Focused.Prompt = lipgloss.NewStyle().Foreground(colorCyan).Bold(true)
	styles.Blurred.Prompt = styles.Focused.Prompt
	styles.Focused.Text = lipgloss.NewStyle().Foreground(colorText)
	styles.Blurred.Text = styles.Focused.Text
	ti.SetStyles(styles)

	now := time.Now()
	m := model{
		address:   addr,
		input:     ti,
		follow:    true,
		gen:       1,
		attempt:   1,
		ticking:   true, // Init() starts the first frame tick
		phase:     phaseLoading,
		now:       now,
		phaseAt:   now,
		connStart: now,
		loaderT0:  now,
	}
	m.addLog(kindSystem, cfgNote)
	m.addLog(kindSystem, fmt.Sprintf("Connecting to TAP Server at %s...", addr))
	return m
}
