package util

import (
	"fmt"
	"os"

	"tap-server/src/model"
)

// ANSI Color Palette for Terminal Logging
const (
	colorReset   = "\033[0m"
	colorBold    = "\033[1m"
	colorDim     = "\033[90m"
	colorCyan    = "\033[36m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorRed     = "\033[31m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
)

// PrintPrettyLog formats and colorizes interaction records for server terminal output
func PrintPrettyLog(r model.InteractionRecord) {
	timeStr := r.Timestamp.Format("15:04:05.000")

	// Level badge styling
	levelColor := colorGreen
	levelBadge := " INFO "
	switch r.Level {
	case "WARN":
		levelColor = colorYellow
		levelBadge = " WARN "
	case "ERROR":
		levelColor = colorRed
		levelBadge = "ERROR "
	}

	actor := r.Actor
	if actor == "" {
		actor = "server"
	}
	remote := r.Remote
	if remote == "" {
		remote = "internal"
	}

	// Compile details payload
	details := ""
	if r.Event != "" {
		details += fmt.Sprintf("event=%s ", r.Event)
	}
	if r.Command != "" {
		details += fmt.Sprintf("cmd=%s ", r.Command)
	}
	if r.Scope != "" {
		details += fmt.Sprintf("scope=%s ", r.Scope)
	}
	if r.Target != "" {
		details += fmt.Sprintf("target=%s ", r.Target)
	}
	if r.Payload != "" {
		details += fmt.Sprintf("payload=%s", r.Payload)
	}

	// Format: HH:MM:SS.ms | LEVEL | TYPE | ACTOR @ REMOTE -> DETAILS
	out := fmt.Sprintf("%s%s%s | %s[%s]%s | %s[%s]%s | %s%s%s@%s%s%s%s -> %s",
		colorDim, timeStr, colorReset,
		levelColor, levelBadge, colorReset,
		colorMagenta, r.Type, colorReset,
		colorBold, colorBlue, actor, colorReset, colorDim, remote,
		colorReset, details,
	)

	if r.Level == "ERROR" {
		fmt.Fprintln(os.Stderr, out)
	} else {
		fmt.Fprintln(os.Stdout, out)
	}
}