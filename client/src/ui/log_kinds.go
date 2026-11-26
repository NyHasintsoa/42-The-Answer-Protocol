package ui

import "tap-client/src/ui/animation"

// --- Log entries ------------------------------------------------------------------

type logKind int

const (
	kindSystem logKind = iota
	kindClient
	kindOK
	kindEvent
	kindServerErr
	kindError
)

// kindStyle returns the two-character gutter label of a log kind and the
// colours of that label and of the message text.
//
// Parameters:
//
//	k: kind that determines the label and palette.
//
// Returns:
//
//	The gutter label, its color, and the message text color.
func kindStyle(k logKind) (label string, labelC, textC animation.RGB) {
	switch k {
	case kindClient:
		return "C:", animation.RGBCyan, animation.RGBText
	case kindOK:
		return "S:", animation.RGBGreen, animation.RGBGreen
	case kindEvent:
		return "S:", animation.RGBMagenta, animation.RGBMagenta
	case kindServerErr:
		return "S:", animation.RGBRed, animation.RGBRed
	case kindError:
		return "!:", animation.RGBRed, animation.RGBRed
	default:
		return "i:", animation.RGBYellow, animation.RGBYellow
	}
}
