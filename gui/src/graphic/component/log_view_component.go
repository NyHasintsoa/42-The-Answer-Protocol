package component

import (
	"tap-gui/src/utils"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type LogViewComponent struct {
	Logs []string
}

func NewLogViewComponent() *LogViewComponent {
	return &LogViewComponent{
		Logs: []string{
			"Cras mattis consectetur purus sit amet fermentum. Cras justo odio, dapibus ac facilisis in, egestas eget quam. Morbi leo risus, porta ac consectetur ac, vestibulum at eros.",
			"Praesent commodo cursus magna, vel scelerisque nisl consectetur et. Vivamus sagittis lacus vel augue laoreet rutrum faucibus dolor auctor.",
			"Aenean lacinia bibendum nulla sed consectetur. Praesent commodo cursus magna, vel scelerisque nisl consectetur et. Donec sed odio dui. Donec ullamcorper nulla non metus auctor fringilla.",
			"Cras mattis consectetur purus sit amet fermentum. Cras justo odio, dapibus ac facilisis in, egestas eget quam. Morbi leo risus, porta ac consectetur ac, vestibulum at eros.",
		},
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
	for _, logMsg := range lvc.Logs {
		if logY > rect.Y+rect.Height-25 {
			break
		}
		drawTextWrapped(
			logMsg,
			int32(rect.X+14),
			int32(logY),
			int32(rect.Width-28),
			10,
			utils.ThemeTextMuted,
		)
		logY += 45
	}
}
