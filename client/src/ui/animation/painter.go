package animation

import "image/color"

// Painter fades colours towards the background, so a whole screen can
// materialise or dissolve without touching the individual styles.
type Painter struct{ Fade float64 }

// R fades c from the background color according to the receiver's Fade value.
//
// Parameters:
//
//	c: RGB color to fade.
//
// Returns:
//
//	The faded color as a standard-library color.
func (p Painter) R(c RGB) color.Color {
	if p.Fade >= 1 {
		return c.ToColor()
	}
	return RGBBg.Mix(c, p.Fade).ToColor()
}

// C converts col to RGB and fades it using the receiver's Fade value.
//
// Parameters:
//
//	col: standard-library color to fade.
//
// Returns:
//
//	The faded color as a standard-library color.
func (p Painter) C(col color.Color) color.Color { return p.R(toRGB(col)) }
