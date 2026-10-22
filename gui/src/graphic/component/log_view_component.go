package component

import (
	"tap-gui/src/graphic/component/ui"
	"tap-gui/src/utils"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type LogViewComponent struct {
	Logs         []string
	ScrollY      float32
	lastLogCount int
	followTail   bool
}

// NewLogViewComponent creates an empty log view that follows new messages.
func NewLogViewComponent() *LogViewComponent {
	return &LogViewComponent{
		Logs:       []string{},
		followTail: true,
	}
}

// AddLog appends msg to the log displayed by the component.
func (lvc *LogViewComponent) AddLog(msg string) {
	lvc.Logs = append(lvc.Logs, msg)
}

// Render draws the log messages within rect.
// rect is the component's screen-space bounds.
func (lvc *LogViewComponent) Render(rect rl.Rectangle) {
	rl.DrawRectangleRounded(rect, 0.02, 6, utils.ThemeSurface)
	rl.DrawRectangleRoundedLinesEx(rect, 0.02, 6, 2, utils.ThemeBorder)

	title := "LOG VIEW"
	rl.DrawText(title, int32(rect.X+14), int32(rect.Y+12), 14, utils.ThemeText)
	rl.DrawLineEx(rl.NewVector2(rect.X+10, rect.Y+32), rl.NewVector2(rect.X+rect.Width-10, rect.Y+32), 1, utils.ThemeBorder)

	viewport := rl.NewRectangle(rect.X+8, rect.Y+38, rect.Width-16, rect.Height-46)
	contentHeights := make([]float32, len(lvc.Logs))
	contentHeight := float32(0)
	for i, logMsg := range lvc.Logs {
		lines := ui.WrapTextLines(logMsg, 10, int32(viewport.Width-14))
		contentHeights[i] = float32(len(lines)*12 + 6)
		contentHeight += contentHeights[i]
	}
	maxScroll := contentHeight - viewport.Height
	if maxScroll < 0 {
		maxScroll = 0
	}
	if lvc.followTail && len(lvc.Logs) > lvc.lastLogCount {
		lvc.ScrollY = maxScroll
	}
	previousScroll := lvc.ScrollY
	maxScroll = ui.UpdateScrollOffset(&lvc.ScrollY, viewport, contentHeight)
	if lvc.ScrollY != previousScroll {
		lvc.followTail = lvc.ScrollY >= maxScroll
	}
	lvc.lastLogCount = len(lvc.Logs)

	rl.BeginScissorMode(int32(viewport.X), int32(viewport.Y), int32(viewport.Width), int32(viewport.Height))
	logY := viewport.Y - lvc.ScrollY
	for i, logMsg := range lvc.Logs {
		ui.DrawTextWrapped(
			logMsg,
			int32(viewport.X+4),
			int32(logY),
			int32(viewport.Width-14),
			15,
			utils.ThemeTextMuted,
		)
		logY += contentHeights[i]
	}
	rl.EndScissorMode()
	ui.DrawScrollIndicator(viewport, contentHeight, lvc.ScrollY)
}
