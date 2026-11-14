package utils

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	// ThemeBackground is the default color behind the interface.
	ThemeBackground = rl.NewColor(7, 12, 20, 255)
	// ThemeSurface is the base color for interface panels.
	ThemeSurface = rl.NewColor(15, 23, 36, 255)
	// ThemeSurfaceRaised is the color for panels raised above the base surface.
	ThemeSurfaceRaised = rl.NewColor(22, 33, 50, 255)
	// ThemeSurfaceInset is the color for inset interface regions.
	ThemeSurfaceInset = rl.NewColor(10, 17, 28, 255)
	// ThemeBorder is the standard interface border color.
	ThemeBorder = rl.NewColor(47, 67, 92, 255)
	// ThemeBorderStrong is the emphasized interface border color.
	ThemeBorderStrong = rl.NewColor(0, 184, 212, 255)
	// ThemeAccent is the primary highlight color.
	ThemeAccent = rl.NewColor(0, 184, 212, 255)
	// ThemeAccentHover is the highlight color used while hovering.
	ThemeAccentHover = rl.NewColor(38, 211, 225, 255)
	// ThemeAccentPressed is the highlight color used while pressing.
	ThemeAccentPressed = rl.NewColor(0, 137, 160, 255)
	// ThemeText is the standard interface text color.
	ThemeText = rl.NewColor(232, 241, 247, 255)
	// ThemeTextMuted is the secondary interface text color.
	ThemeTextMuted = rl.NewColor(157, 177, 194, 255)
	// ThemeTextSubtle is the subdued interface text color.
	ThemeTextSubtle = rl.NewColor(105, 128, 149, 255)
	// ThemeDanger is the standard color for destructive actions.
	ThemeDanger = rl.NewColor(190, 58, 72, 255)
	// ThemeDangerHover is the color for destructive actions while hovered.
	ThemeDangerHover = rl.NewColor(225, 78, 91, 255)
	// ThemeSuccess is the standard color for successful states.
	ThemeSuccess = rl.NewColor(63, 194, 132, 255)
)
