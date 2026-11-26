package theme

import "time"

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
