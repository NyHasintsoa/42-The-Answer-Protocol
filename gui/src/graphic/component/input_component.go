package component

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

type InputComponent struct {
	Rect        rl.Rectangle
	Text        string
	Placeholder string
	IsFocused   bool
	MaxLength   int
	cursorTimer float32
	showCursor  bool
}

func NewInputComponent(x, y, w, h float32, placeholder string, maxLen int) *InputComponent {
	return &InputComponent{
		Rect:        rl.NewRectangle(x, y, w, h),
		Placeholder: placeholder,
		MaxLength:   maxLen,
		showCursor:  true,
	}
}

func (ic *InputComponent) Update(dt float32) bool {
	mousePos := rl.GetMousePosition()

	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		ic.IsFocused = rl.CheckCollisionPointRec(mousePos, ic.Rect)
	}

	if !ic.IsFocused {
		return false
	}

	ic.cursorTimer += dt
	if ic.cursorTimer >= 0.5 {
		ic.showCursor = !ic.showCursor
		ic.cursorTimer = 0
	}

	char := rl.GetCharPressed()
	for char > 0 {
		if len(ic.Text) < ic.MaxLength && char >= 32 && char <= 125 {
			ic.Text += string(char)
		}
		char = rl.GetCharPressed()
	}

	if rl.IsKeyPressed(rl.KeyBackspace) && len(ic.Text) > 0 {
		ic.Text = ic.Text[:len(ic.Text)-1]
	}

	if rl.IsKeyPressed(rl.KeyEnter) || rl.IsKeyPressed(rl.KeyKpEnter) {
		return true
	}

	return false
}

func (ic *InputComponent) Render() {
	bgColor := rl.NewColor(24, 28, 42, 255)
	borderColor := rl.NewColor(50, 70, 100, 255)
	if ic.IsFocused {
		borderColor = rl.NewColor(0, 180, 255, 255)
	}

	rl.DrawRectangleRounded(ic.Rect, 0.2, 8, bgColor)
	rl.DrawRectangleRoundedLinesEx(ic.Rect, 0.2, 8, 2, borderColor)

	paddingX := float32(16)
	posY := ic.Rect.Y + (ic.Rect.Height-16)/2

	if len(ic.Text) == 0 && !ic.IsFocused {
		rl.DrawText(ic.Placeholder, int32(ic.Rect.X+paddingX), int32(posY), 16, rl.NewColor(120, 135, 160, 255))
	} else {
		rl.DrawText(ic.Text, int32(ic.Rect.X+paddingX), int32(posY), 16, rl.RayWhite)

		if ic.IsFocused && ic.showCursor {
			textWidth := float32(rl.MeasureText(ic.Text, 16))
			cursorX := ic.Rect.X + paddingX + textWidth + 2
			rl.DrawLineEx(
				rl.NewVector2(cursorX, posY-2),
				rl.NewVector2(cursorX, posY+18),
				2,
				rl.NewColor(0, 180, 255, 255),
			)
		}
	}
}

func (ic *InputComponent) GetAndClearText() string {
	t := ic.Text
	ic.Text = ""
	return t
}