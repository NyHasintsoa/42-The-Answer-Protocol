package component

import (
	"image/color"
	"math"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
)

// --- Small math / colour helpers ----------------------------------------------

type rgb struct{ r, g, b float64 }

func toRGB(c color.Color) rgb {
	r, g, b, _ := c.RGBA()
	return rgb{float64(r >> 8), float64(g >> 8), float64(b >> 8)}
}

func (a rgb) mix(b rgb, t float64) rgb {
	t = clamp01(t)
	return rgb{a.r + (b.r-a.r)*t, a.g + (b.g-a.g)*t, a.b + (b.b-a.b)*t}
}

func (a rgb) toColor() color.Color {
	return color.RGBA{R: channel(a.r), G: channel(a.g), B: channel(a.b), A: 0xFF}
}

func channel(v float64) uint8 {
	return uint8(math.Round(math.Max(0, math.Min(255, v))))
}

// painter fades colours towards the background, so a whole screen can
// materialise or dissolve without touching the individual styles.
type painter struct{ fade float64 }

func (p painter) r(c rgb) color.Color {
	if p.fade >= 1 {
		return c.toColor()
	}
	return rgbBg.mix(c, p.fade).toColor()
}

func (p painter) c(col color.Color) color.Color { return p.r(toRGB(col)) }

func clamp01(x float64) float64 { return math.Max(0, math.Min(1, x)) }

func clampInt(v, lo, hi int) int { return max(lo, min(v, hi)) }

func easeOutCubic(t float64) float64 {
	u := 1 - clamp01(t)
	return 1 - u*u*u
}

func easeInCubic(t float64) float64 {
	t = clamp01(t)
	return t * t * t
}

func easeInOutSine(t float64) float64 { return -(math.Cos(math.Pi*clamp01(t)) - 1) / 2 }

func smoothstep(t float64) float64 {
	t = clamp01(t)
	return t * t * (3 - 2*t)
}

// fraction returns d/total as a float (1 when total is zero).
func fraction(d, total time.Duration) float64 {
	if total <= 0 {
		return 1
	}
	return float64(d) / float64(total)
}

// gradient is a looping cyan -> blue -> magenta ramp (u wraps around).
func gradient(u float64) rgb {
	stops := [...]rgb{rgbCyan, rgbBlue, rgbMagenta}
	u -= math.Floor(u)
	seg := u * float64(len(stops))
	i := int(seg)
	return stops[i%len(stops)].mix(stops[(i+1)%len(stops)], smoothstep(seg-float64(i)))
}

// shimmerColor is the gradient plus a bright diagonal band that sweeps across.
func shimmerColor(col, row, width int, t float64) rgb {
	base := gradient(float64(col)/float64(max(width, 1))*0.8 + t*0.10)
	const period = 2.8
	pos := -8 + (math.Mod(t, period)/period)*(float64(width)+16)*1.5
	d := float64(col) + float64(row)*1.6 - pos
	return base.mix(rgbWhite, math.Exp(-d*d/(2*3.5*3.5))*0.85)
}

// shimmerLine renders s with one shimmering colour per rune.
func shimmerLine(s string, row int, t float64, p painter) string {
	runes := []rune(s)
	var b strings.Builder
	for col, r := range runes {
		if r == ' ' {
			b.WriteByte(' ')
			continue
		}
		c := shimmerColor(col, row, len(runes), t)
		if strings.ContainsRune("╗╔╝╚║═", r) {
			c = c.mix(rgbBg, 0.55) // the "shadow" of the block letters
		}
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(p.r(c)).Render(string(r)))
	}
	return b.String()
}
