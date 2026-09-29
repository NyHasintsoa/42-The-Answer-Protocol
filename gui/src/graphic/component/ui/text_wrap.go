package ui

import (
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func DrawTextWrapped(text string, x, y, maxWidth int32, fontSize int32, color rl.Color) int32 {
	lineHeight := fontSize + 2
	currentY := y

	for paragraph := range strings.SplitSeq(text, "\n") {
		line := ""
		for word := range strings.FieldsSeq(paragraph) {
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
