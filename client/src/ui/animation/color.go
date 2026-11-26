package animation

import (
	"image/color"
	"math"

	"tap-client/src/ui/theme"
)

// RGB is a colour with float channels in the [0, 255] range, so animations can
// blend colours without touching individual lipgloss styles.
type RGB struct{ R, G, B float64 }

// toRGB converts c's 8-bit RGB channels into floating-point channels.
//
// Parameters:
//
//	c: source color to convert.
//
// Returns:
//
//	The color represented as RGB channels in the range [0, 255].
func toRGB(c color.Color) RGB {
	r, g, b, _ := c.RGBA()
	return RGB{float64(r >> 8), float64(g >> 8), float64(b >> 8)}
}

// Mix blends a with b by the clamped fraction t.
//
// Parameters:
//
//	b: destination color.
//	t: blend fraction, clamped to [0, 1].
//
// Returns:
//
//	The interpolated RGB color.
func (a RGB) Mix(b RGB, t float64) RGB {
	t = clamp01(t)
	return RGB{a.R + (b.R-a.R)*t, a.G + (b.G-a.G)*t, a.B + (b.B-a.B)*t}
}

// ToColor converts the receiver to an opaque standard-library color.
//
// Returns:
//
//	The receiver represented as color.RGBA.
func (a RGB) ToColor() color.Color {
	return color.RGBA{R: channel(a.R), G: channel(a.G), B: channel(a.B), A: 0xFF}
}

// channel clamps v to the byte color range and rounds it to the nearest value.
//
// Parameters:
//
//	v: floating-point color channel.
//
// Returns:
//
//	The channel represented as an 8-bit value.
func channel(v float64) uint8 {
	return uint8(math.Round(math.Max(0, math.Min(255, v))))
}

// clamp01 limits x to the closed interval [0, 1].
//
// Parameters:
//
//	x: value to clamp.
//
// Returns:
//
//	x clamped to [0, 1].
func clamp01(x float64) float64 { return math.Max(0, math.Min(1, x)) }

// Clamp01 clamps x to the [0, 1] range.
//
// Parameters:
//
//	x: value to clamp.
//
// Returns:
//
//	x clamped to [0, 1].
func Clamp01(x float64) float64 { return clamp01(x) }

// ClampInt limits v to the inclusive interval from lo to hi.
//
// Parameters:
//
//	v: value to clamp.
//	lo: lower bound of the interval.
//	hi: upper bound of the interval.
//
// Returns:
//
//	v constrained to [lo, hi].
func ClampInt(v, lo, hi int) int { return max(lo, min(v, hi)) }

// The same colours as the theme palette, as plain RGB, so animations can blend
// them.
var (
	RGBBg      = toRGB(theme.ColorBg)
	RGBSurface = toRGB(theme.ColorSurface)
	RGBCyan    = toRGB(theme.ColorCyan)
	RGBGreen   = toRGB(theme.ColorGreen)
	RGBYellow  = toRGB(theme.ColorYellow)
	RGBRed     = toRGB(theme.ColorRed)
	RGBMagenta = toRGB(theme.ColorMagenta)
	RGBBlue    = toRGB(theme.ColorBlue)
	RGBText    = toRGB(theme.ColorText)
	RGBMuted   = toRGB(theme.ColorMuted)
	RGBBorder  = toRGB(theme.ColorBorder)
	RGBWhite   = toRGB(theme.ColorWhite)
)
