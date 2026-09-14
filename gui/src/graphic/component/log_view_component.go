package component

import (
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
	// Container Panel
	rl.DrawRectangleRounded(rect, 0.02, 6, rl.NewColor(245, 247, 250, 255))
	rl.DrawRectangleRoundedLinesEx(rect, 0.02, 6, 2, rl.NewColor(180, 190, 200, 255))

	// Header
	title := "LOG VIEW"
	rl.DrawText(title, int32(rect.X+14), int32(rect.Y+12), 14, rl.NewColor(40, 50, 60, 255))
	rl.DrawLineEx(rl.NewVector2(rect.X+10, rect.Y+32), rl.NewVector2(rect.X+rect.Width-10, rect.Y+32), 1, rl.NewColor(210, 220, 230, 255))

	// Wrapped log entries
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
			rl.NewColor(60, 70, 85, 255),
		)
		logY += 45
	}
}
