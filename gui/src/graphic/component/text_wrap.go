package component

import (
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func drawTextWrapped(text string, x, y, maxWidth int32, fontSize int32, color rl.Color) int32 {
	lineHeight := fontSize + 2
	currentY := y

	for _, paragraph := range strings.Split(text, "\n") {
		line := ""
		for _, word := range strings.Fields(paragraph) {
			candidate := word
			if line != "" {
				candidate = line + " " + word
			}

			if line != "" && rl.MeasureText(candidate, fontSize) > maxWidth {
				rl.DrawText(line, x, currentY, fontSize, color)
				currentY += lineHeight
				line = word
				continue
			}
			line = candidate
		}

		if line != "" {
			rl.DrawText(line, x, currentY, fontSize, color)
		}
		currentY += lineHeight
	}

	return currentY
}
