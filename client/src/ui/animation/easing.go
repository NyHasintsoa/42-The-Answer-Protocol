package animation

import (
	"math"
	"time"
)

// Fraction returns d/total as a float (one when total is non-positive).
//
// Parameters:
//
//	d: elapsed duration.
//	total: duration that represents a complete transition.
//
// Returns:
//
//	The elapsed fraction, or one when total is non-positive.
func Fraction(d, total time.Duration) float64 {
	if total <= 0 {
		return 1
	}
	return float64(d) / float64(total)
}

// EaseOutCubic applies cubic ease-out to a normalized progress value.
//
// Parameters:
//
//	t: progress value, clamped to [0, 1].
//
// Returns:
//
//	Eased progress in [0, 1].
func EaseOutCubic(t float64) float64 {
	u := 1 - clamp01(t)
	return 1 - u*u*u
}

// EaseInCubic applies cubic ease-in to a normalized progress value.
//
// Parameters:
//
//	t: progress value, clamped to [0, 1].
//
// Returns:
//
//	Eased progress in [0, 1].
func EaseInCubic(t float64) float64 {
	t = clamp01(t)
	return t * t * t
}

// EaseInOutSine applies a sine ease-in/ease-out curve to normalized progress.
//
// Parameters:
//
//	t: progress value, clamped to [0, 1].
//
// Returns:
//
//	Eased progress in [0, 1].
func EaseInOutSine(t float64) float64 { return -(math.Cos(math.Pi*clamp01(t)) - 1) / 2 }

// smoothstep applies the cubic smoothstep curve to normalized progress.
//
// Parameters:
//
//	t: progress value, clamped to [0, 1].
//
// Returns:
//
//	Eased progress in [0, 1].
func smoothstep(t float64) float64 {
	t = clamp01(t)
	return t * t * (3 - 2*t)
}

// gradient is a looping cyan -> blue -> magenta ramp (u wraps around).
//
// Parameters:
//
//	u: position in the repeating color gradient.
//
// Returns:
//
//	The interpolated gradient color at u.
func gradient(u float64) RGB {
	stops := [...]RGB{RGBCyan, RGBBlue, RGBMagenta}
	u -= math.Floor(u)
	seg := u * float64(len(stops))
	i := int(seg)
	return stops[i%len(stops)].Mix(stops[(i+1)%len(stops)], smoothstep(seg-float64(i)))
}
