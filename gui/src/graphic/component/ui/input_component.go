package ui

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
	MinHeight   float32
	MaxHeight   float32
}

func NewInputComponent(x, y, w, h float32, placeholder string, maxLen int) *InputComponent {
	if h <= 0 {
		h = 48
	}
	return &InputComponent{
		Rect:        rl.NewRectangle(x, y, w, h),
		Placeholder: placeholder,
		MaxLength:   maxLen,
		showCursor:  true,
		MinHeight:   h,
		MaxHeight:   120,
	}
}

func (ic *InputComponent) RecalculateHeight() {
	fontSize := int32(16)
	paddingX := float32(16)
	paddingY := float32(12)
	lineHeight := float32(fontSize + 4)

	textMaxWidth := int32(ic.Rect.Width - (paddingX * 2))
	if textMaxWidth <= 0 {
		return
	}

	lines := WrapTextLines(ic.Text, fontSize, textMaxWidth)
	calcH := (paddingY * 2) + float32(len(lines))*lineHeight

	if calcH < ic.MinHeight {
		calcH = ic.MinHeight
	}
	if ic.MaxHeight > 0 && calcH > ic.MaxHeight {
		calcH = ic.MaxHeight
	}

	ic.Rect.Height = calcH
}

func (ic *InputComponent) Update(dt float32) bool {
	mousePos := rl.GetMousePosition()

	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		ic.IsFocused = rl.CheckCollisionPointRec(mousePos, ic.Rect)
	}

	if !ic.IsFocused {
		ic.RecalculateHeight()
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

	ic.RecalculateHeight()

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

	rl.DrawRectangleRounded(ic.Rect, 0.15, 8, bgColor)
	rl.DrawRectangleRoundedLinesEx(ic.Rect, 0.15, 8, 2, borderColor)

	fontSize := int32(16)
	paddingX := float32(16)
	paddingY := float32(12)
	lineHeight := float32(fontSize + 4)
	textMaxWidth := int32(ic.Rect.Width - (paddingX * 2))

	if len(ic.Text) == 0 && !ic.IsFocused {
		posY := ic.Rect.Y + (ic.Rect.Height-float32(fontSize))/2
		rl.DrawText(ic.Placeholder, int32(ic.Rect.X+paddingX), int32(posY), fontSize, rl.NewColor(120, 135, 160, 255))
		return
	}

	lines := WrapTextLines(ic.Text, fontSize, textMaxWidth)
	visibleLines := int((ic.Rect.Height - (paddingY * 2)) / lineHeight)
	if visibleLines < 1 {
		visibleLines = 1
	}

	startIdx := 0
	if len(lines) > visibleLines {
		startIdx = len(lines) - visibleLines
	}

	currY := ic.Rect.Y + paddingY
	for i := startIdx; i < len(lines); i++ {
		rl.DrawText(lines[i], int32(ic.Rect.X+paddingX), int32(currY), fontSize, rl.RayWhite)
		currY += lineHeight
	}

	if ic.IsFocused && ic.showCursor {
		lastLine := lines[len(lines)-1]
		lastLineW := float32(rl.MeasureText(lastLine, fontSize))
		lastLineIdx := len(lines) - 1 - startIdx
		cursorY := ic.Rect.Y + paddingY + float32(lastLineIdx)*lineHeight
		cursorX := ic.Rect.X + paddingX + lastLineW + 2

		rl.DrawLineEx(
			rl.NewVector2(cursorX, cursorY),
			rl.NewVector2(cursorX, cursorY+float32(fontSize)),
			2,
			rl.NewColor(0, 180, 255, 255),
		)
	}
}

func (ic *InputComponent) GetAndClearText() string {
	t := ic.Text
	ic.Text = ""
	ic.RecalculateHeight()
	return t
}
