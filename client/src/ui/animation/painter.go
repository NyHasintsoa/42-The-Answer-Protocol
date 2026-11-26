package animation

import "image/color"

// painter fades colours towards the background, so a whole screen can
// materialise or dissolve without touching the individual styles.
type Painter struct{ Fade float64 }

func (p Painter) R(c RGB) color.Color {
	if p.Fade >= 1 {
		return c.ToColor()
	}
	return RGBBg.Mix(c, p.Fade).ToColor()
}

func (p Painter) C(col color.Color) color.Color { return p.R(toRGB(col)) }
