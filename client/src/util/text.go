package util

import (
	"errors"
	"io"
	"net"
	"strings"
	"unicode"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// --- Text helpers -------------------------------------------------------------

// Sanitize makes server text safe to print: tabs become spaces and every other
// control character (ESC, BEL, CR...) is dropped, so a server or another
// player can never inject terminal escape sequences into the log.
//
// Parameters:
//
//	s: untrusted text received from the server.
//
// Returns:
//
//	Text with tabs expanded and other control characters removed.
func Sanitize(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r == '\t':
			b.WriteString("    ")
		case unicode.IsControl(r):
			// dropped
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// runeWidth returns the terminal cell width of a single rune.
//
// Parameters:
//
//	r: rune whose display width is measured.
//
// Returns:
//
//	The number of terminal cells occupied by r.
func runeWidth(r rune) int { return lipgloss.Width(string(r)) }

// WrapText splits plain text into lines of at most width terminal cells. It
// prefers to break on spaces and hard-breaks words that are too long.
//
// Parameters:
//
//	s: plain text to wrap.
//	width: maximum terminal-cell width per line; values below one are treated as one.
//
// Returns:
//
//	Wrapped lines, including one empty line when s is empty.
func WrapText(s string, width int) []string {
	if width < 1 {
		width = 1
	}

	runes := []rune(s)
	var lines []string
	for len(runes) > 0 {
		used, n, lastSpace := 0, 0, -1
		for n < len(runes) {
			w := runeWidth(runes[n])
			if used+w > width {
				break
			}
			used += w
			if runes[n] == ' ' {
				lastSpace = n
			}
			n++
		}
		if n == len(runes) { // the rest fits
			lines = append(lines, string(runes))
			break
		}
		cut, skip := n, 0
		switch {
		case n == 0: // one rune wider than the whole line: still make progress
			cut = 1
		case lastSpace > 0:
			cut, skip = lastSpace, 1 // break on the space and drop it
		}
		lines = append(lines, string(runes[:cut]))
		runes = runes[cut+skip:]
		for len(runes) > 0 && runes[0] == ' ' { // no blanks at the start of a continuation line
			runes = runes[1:]
		}
	}
	if len(lines) == 0 {
		lines = []string{""}
	}
	return lines
}

// TruncateCells truncates s to at most w terminal cells.
//
// Parameters:
//
//	s: text to truncate.
//	w: maximum terminal-cell width.
//
// Returns:
//
//	The truncated text with an ellipsis, or an empty string when w is non-positive.
func TruncateCells(s string, w int) string {
	if w <= 0 {
		return ""
	}
	return ansi.Truncate(s, w, "…")
}

// PadTo right-pads s with spaces up to w terminal cells.
//
// Parameters:
//
//	s: text to pad.
//	w: minimum terminal-cell width.
//
// Returns:
//
//	s padded to at least w cells; text wider than w is returned unchanged.
func PadTo(s string, w int) string {
	if gap := w - lipgloss.Width(s); gap > 0 {
		return s + strings.Repeat(" ", gap)
	}
	return s
}

// Center pads s with spaces so it is centered within w terminal cells.
//
// Parameters:
//
//	s: text to center.
//	w: target terminal-cell width.
//
// Returns:
//
//	s centered and padded when it is narrower than w, or unchanged otherwise.
func Center(s string, w int) string {
	sw := lipgloss.Width(s)
	if sw >= w {
		return s
	}
	left := (w - sw) / 2
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", w-sw-left)
}

// SpaceOut inserts a single space between every rune of s.
//
// Parameters:
//
//	s: text whose runes are separated.
//
// Returns:
//
//	The runes of s joined with single spaces.
func SpaceOut(s string) string {
	var b strings.Builder
	for i, r := range []rune(s) {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteRune(r)
	}
	return b.String()
}

// DescribeErr turns a network error into something short enough for the
// loader.
//
// Parameters:
//
//	err: non-nil network error to describe.
//
// Returns:
//
//	A concise description of err suitable for the connection loader.
func DescribeErr(err error) string {
	var opErr *net.OpError
	switch {
	case errors.Is(err, io.EOF):
		return "server closed the connection"
	case errors.As(err, &opErr) && opErr.Err != nil:
		return opErr.Err.Error()
	default:
		return err.Error()
	}
}
