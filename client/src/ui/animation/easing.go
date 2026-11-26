package animation

import (
	"math"
	"time"
)

// fraction returns d/total as a float (1 when total is zero).
func Fraction(d, total time.Duration) float64 {
	if total <= 0 {
		return 1
	}
	return float64(d) / float64(total)
}

func EaseOutCubic(t float64) float64 {
	u := 1 - clamp01(t)
	return 1 - u*u*u
}

func EaseInCubic(t float64) float64 {
	t = clamp01(t)
	return t * t * t
}

func EaseInOutSine(t float64) float64 { return -(math.Cos(math.Pi*clamp01(t)) - 1) / 2 }

func smoothstep(t float64) float64 {
	t = clamp01(t)
	return t * t * (3 - 2*t)
}

// gradient is a looping cyan -> blue -> magenta ramp (u wraps around).
func gradient(u float64) RGB {
	stops := [...]RGB{RGBCyan, RGBBlue, RGBMagenta}
	u -= math.Floor(u)
	seg := u * float64(len(stops))
	i := int(seg)
	return stops[i%len(stops)].Mix(stops[(i+1)%len(stops)], smoothstep(seg-float64(i)))
}
