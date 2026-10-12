package ui

import (
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func WrapTextLines(text string, fontSize int32, maxWidth int32) []string {
	var lines []string
	if maxWidth <= 0 {
		return []string{text}
	}

	paragraphs := strings.Split(text, "\n")
	for _, paragraph := range paragraphs {
		if paragraph == "" {
			lines = append(lines, "")
			continue
		}

		words := strings.Fields(paragraph)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}

		currentLine := ""

		for _, word := range words {
			// If a single word/sequence is longer than maxWidth, break character-by-character
			if rl.MeasureText(word, fontSize) > maxWidth {
				if currentLine != "" {
					lines = append(lines, currentLine)
					currentLine = ""
				}

				subWord := ""
				for _, ch := range word {
					candidate := subWord + string(ch)
					if rl.MeasureText(candidate, fontSize) > maxWidth {
						if subWord != "" {
							lines = append(lines, subWord)
						}
						subWord = string(ch)
					} else {
						subWord = candidate
					}
				}
				if subWord != "" {
					currentLine = subWord
				}
				continue
			}

			candidate := word
			if currentLine != "" {
				candidate = currentLine + " " + word
			}

			if rl.MeasureText(candidate, fontSize) > maxWidth {
				if currentLine != "" {
					lines = append(lines, currentLine)
				}
				currentLine = word
			} else {
				currentLine = candidate
			}
		}

		if currentLine != "" {
			lines = append(lines, currentLine)
		}
	}

	if len(lines) == 0 {
		lines = append(lines, "")
	}

	return lines
}

func DrawTextWrapped(text string, x, y, maxWidth int32, fontSize int32, color rl.Color) int32 {
	lineHeight := fontSize + 2
	lines := WrapTextLines(text, fontSize, maxWidth)
	currentY := y

	for _, line := range lines {
		if color.A > 0 {
			rl.DrawText(line, x, currentY, fontSize, color)
		}
		currentY += lineHeight
	}

	return currentY
}
