package component

import (
	"tap-gui/src/graphic/component/ui"
	"tap-gui/src/utils"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type LogViewComponent struct {
	Logs []string
}

func NewLogViewComponent() *LogViewComponent {
	return &LogViewComponent{
		Logs: []string{},
	}
}

func (lvc *LogViewComponent) AddLog(msg string) {
	lvc.Logs = append(lvc.Logs, msg)
}

func (lvc *LogViewComponent) Render(rect rl.Rectangle) {
	rl.DrawRectangleRounded(rect, 0.02, 6, utils.ThemeSurface)
	rl.DrawRectangleRoundedLinesEx(rect, 0.02, 6, 2, utils.ThemeBorder)

	title := "LOG VIEW"
	rl.DrawText(title, int32(rect.X+14), int32(rect.Y+12), 14, utils.ThemeText)
	rl.DrawLineEx(rl.NewVector2(rect.X+10, rect.Y+32), rl.NewVector2(rect.X+rect.Width-10, rect.Y+32), 1, utils.ThemeBorder)

	logY := rect.Y + 40
	lineHeight := float32(45)
	visibleLogs := int((rect.Height - 50) / lineHeight)
	startIndex := len(lvc.Logs) - visibleLogs
	if startIndex < 0 {
		startIndex = 0
	}
	for _, logMsg := range lvc.Logs[startIndex:] {
		ui.DrawTextWrapped(
			logMsg,
			int32(rect.X+14),
			int32(logY),
			int32(rect.Width-28),
			10,
			utils.ThemeTextMuted,
		)
		logY += lineHeight
	}
}
