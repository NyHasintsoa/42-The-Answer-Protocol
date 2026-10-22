// Package ui contains reusable controls for the game's graphical interface.
package ui

import (
	"tap-gui/src/utils"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var buttonInputEnabled = true

// SetButtonInputEnabled globally enables or disables mouse interaction with buttons.
// enabled determines whether button input is accepted.
func SetButtonInputEnabled(enabled bool) {
	buttonInputEnabled = enabled
}

type Button struct {
	Rect           rl.Rectangle
	Text           string
	BgColor        rl.Color
	HoverBgColor   rl.Color
	HoverTextColor rl.Color
	ClickedColor   rl.Color
	TextColor      rl.Color
	BorderColor    rl.Color
	ShadowColor    rl.Color
	IsClicked      bool
	FontSize       int32
	BorderRadius   float32
	BorderWidth    float32
	Disabled       bool
}

// NewButton creates a button at the given position with the specified size and label.
// x and y are the top-left coordinates; width and height define its size;
// text is the displayed label; fontSize is its text size in pixels.
func NewButton(x, y, width, height float32, text string, fontSize int32) *Button {
	return &Button{
		Rect:           rl.NewRectangle(x, y, width, height),
		Text:           text,
		BgColor:        utils.ThemeSurface,
		HoverBgColor:   utils.ThemeSurfaceRaised,
		ClickedColor:   utils.ThemeAccentPressed,
		TextColor:      utils.ThemeText,
		HoverTextColor: utils.ThemeAccentHover,
		BorderColor:    utils.ThemeAccent,
		ShadowColor:    rl.ColorAlpha(rl.Black, 0.5),
		FontSize:       fontSize,
		BorderRadius:   0.35,
		BorderWidth:    4,
	}
}

// Render draws the button and handles its enabled input state.
func (b *Button) Render() {
	b.IsClicked = false
	currentBorderColor := b.BorderColor
	var currentColor rl.Color

	if b.Disabled || !buttonInputEnabled {
		currentColor = utils.ThemeSurfaceInset
	} else {
		mousePos := rl.GetMousePosition()
		isHovered := rl.CheckCollisionPointRec(mousePos, b.Rect)

		if isHovered {
			currentColor = b.HoverBgColor
		} else {
			currentColor = b.BgColor
		}

		if isHovered && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			b.IsClicked = true
		}

		if isHovered && rl.IsMouseButtonDown(rl.MouseLeftButton) {
			currentColor = b.ClickedColor
		}
	}

	shadowRect := rl.NewRectangle(b.Rect.X+4, b.Rect.Y+6, b.Rect.Width, b.Rect.Height)
	rl.DrawRectangleRounded(shadowRect, b.BorderRadius, 12, b.ShadowColor)

	borderRect := rl.NewRectangle(
		b.Rect.X-b.BorderWidth,
		b.Rect.Y-b.BorderWidth,
		b.Rect.Width+(b.BorderWidth*2),
		b.Rect.Height+(b.BorderWidth*2),
	)
	rl.DrawRectangleRounded(borderRect, b.BorderRadius, 12, currentBorderColor)
	rl.DrawRectangleRounded(b.Rect, b.BorderRadius, 12, currentColor)

	textWidth := float32(rl.MeasureText(b.Text, b.FontSize))
	tx := b.Rect.X + (b.Rect.Width-textWidth)/2
	ty := b.Rect.Y + (b.Rect.Height-float32(b.FontSize))/2

	if currentColor != b.ClickedColor {
		rl.DrawText(b.Text, int32(tx+1), int32(ty+1), b.FontSize, rl.ColorAlpha(rl.White, 0.5))
	}
	rl.DrawText(b.Text, int32(tx), int32(ty), b.FontSize, b.TextColor)
}
