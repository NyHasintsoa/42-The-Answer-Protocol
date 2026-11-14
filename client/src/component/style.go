package component

import (
	"time"

	"charm.land/lipgloss/v2"
)

// --- Tunables ---------------------------------------------------------------

const (
	retryDelay = 2 * time.Second

	// One frame clock drives every animation; it only runs while something moves.
	frameInterval = time.Second / 50

	// Screen transitions.
	loaderFadeIn  = 450 * time.Millisecond // loader card materialises
	successHold   = 450 * time.Millisecond // "Connected" stays visible...
	loaderFadeOut = 180 * time.Millisecond // ...then dissolves
	mainFadeIn    = 350 * time.Millisecond // log screen fades in

	// Log line entrance animation.
	typePerRune    = 5 * time.Millisecond   // typewriter speed...
	typeMinDur     = 140 * time.Millisecond // ...but never instant
	typeMaxDur     = 420 * time.Millisecond // ...and never slow, even for long lines
	lineFadeDur    = 360 * time.Millisecond // text colour fades in from the background
	labelGlowDur   = 520 * time.Millisecond // "C:" / "S:" flashes white, then settles
	lineStagger    = 28 * time.Millisecond  // lines arriving together cascade
	maxStaggerLead = 300 * time.Millisecond // ...but a burst never queues longer than this

	// Log view.
	wheelStep     = 3
	maxLogEntries = 2000
	maxLineRunes  = 4096
	stampLayout   = "15:04:05.000"
	prefixW       = len(stampLayout) + 1 + 2 + 1 // "15:04:05.000 C: "

	// Layout. Also used to hit-test the mouse: screen column of the scrollbar,
	// screen row of the first log line, rows used by everything except the log
	// lines (header 1 + log panel border/title 3 + input panel 4), minimum size.
	barX       = 2
	logTop     = 3
	chromeRows = 8
	minWidth   = 40
	minHeight  = 12
)

// --- Colour palette ---------------------------------------------------------

var (
	colorBg      = lipgloss.Color("#070C14")
	colorSurface = lipgloss.Color("#0F1724")
	colorCyan    = lipgloss.Color("#00B8D4")
	colorGreen   = lipgloss.Color("#3FC284")
	colorYellow  = lipgloss.Color("#F5B041")
	colorRed     = lipgloss.Color("#E14E5B")
	colorMagenta = lipgloss.Color("#C678DD")
	colorBlue    = lipgloss.Color("#61AFEF")
	colorText    = lipgloss.Color("#E8F1F7")
	colorMuted   = lipgloss.Color("#9DB1C2")
	colorBorder  = lipgloss.Color("#2F435C")
	colorWhite   = lipgloss.Color("#FFFFFF")

	// The same colours as plain RGB, so animations can blend them.
	rgbBg      = toRGB(colorBg)
	rgbSurface = toRGB(colorSurface)
	rgbCyan    = toRGB(colorCyan)
	rgbGreen   = toRGB(colorGreen)
	rgbYellow  = toRGB(colorYellow)
	rgbRed     = toRGB(colorRed)
	rgbMagenta = toRGB(colorMagenta)
	rgbBlue    = toRGB(colorBlue)
	rgbText    = toRGB(colorText)
	rgbMuted   = toRGB(colorMuted)
	rgbBorder  = toRGB(colorBorder)
	rgbWhite   = toRGB(colorWhite)
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// logoTAP is the "ANSI Shadow" rendering of TAP (25 columns x 6 rows).
var logoTAP = []string{
	"████████╗ █████╗ ██████╗ ",
	"╚══██╔══╝██╔══██╗██╔══██╗",
	"   ██║   ███████║██████╔╝",
	"   ██║   ██╔══██║██╔═══╝ ",
	"   ██║   ██║  ██║██║     ",
	"   ╚═╝   ╚═╝  ╚═╝╚═╝     ",
}
