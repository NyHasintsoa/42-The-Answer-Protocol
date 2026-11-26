package animation

import (
	"math"
	"strings"

	"charm.land/lipgloss/v2"
)

// shimmerColor is the gradient plus a bright diagonal band that sweeps across.
func shimmerColor(col, row, width int, t float64) RGB {
	base := gradient(float64(col)/float64(max(width, 1))*0.8 + t*0.10)
	const period = 2.8
	pos := -8 + (math.Mod(t, period)/period)*(float64(width)+16)*1.5
	d := float64(col) + float64(row)*1.6 - pos
	return base.Mix(RGBWhite, math.Exp(-d*d/(2*3.5*3.5))*0.85)
}

// ShimmerLine renders s with one shimmering colour per rune.
func ShimmerLine(s string, row int, t float64, p Painter) string {
	runes := []rune(s)
	var b strings.Builder
	for col, r := range runes {
		if r == ' ' {
			b.WriteByte(' ')
			continue
		}
		c := shimmerColor(col, row, len(runes), t)
		if strings.ContainsRune("╗╔╝╚║═", r) {
			c = c.Mix(RGBBg, 0.55) // the "shadow" of the block letters
		}
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(p.R(c)).Render(string(r)))
	}
	return b.String()
}
