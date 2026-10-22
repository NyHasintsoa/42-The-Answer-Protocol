package ui

import rl "github.com/gen2brain/raylib-go/raylib"

func UpdateScrollOffset(offset *float32, viewport rl.Rectangle, contentHeight float32) float32 {
	maxOffset := contentHeight - viewport.Height
	if maxOffset < 0 {
		maxOffset = 0
	}

	mousePosition := rl.GetMousePosition()
	if rl.CheckCollisionPointRec(mousePosition, viewport) {
		*offset -= rl.GetMouseWheelMove() * 36
	}
	if *offset < 0 {
		*offset = 0
	} else if *offset > maxOffset {
		*offset = maxOffset
	}
	return maxOffset
}

func DrawScrollIndicator(viewport rl.Rectangle, contentHeight, offset float32) {
	if contentHeight <= viewport.Height || viewport.Height <= 0 {
		return
	}

	track := rl.NewRectangle(viewport.X+viewport.Width-5, viewport.Y, 3, viewport.Height)
	rl.DrawRectangleRounded(track, 1, 4, rl.NewColor(80, 90, 105, 170))

	thumbHeight := viewport.Height * viewport.Height / contentHeight
	if thumbHeight < 20 {
		thumbHeight = 20
	}
	if thumbHeight > viewport.Height {
		thumbHeight = viewport.Height
	}
	maxOffset := contentHeight - viewport.Height
	thumbY := viewport.Y + (offset/maxOffset)*(viewport.Height-thumbHeight)
	thumb := rl.NewRectangle(track.X, thumbY, track.Width, thumbHeight)
	rl.DrawRectangleRounded(thumb, 1, 4, rl.NewColor(24, 160, 178, 230))
}
