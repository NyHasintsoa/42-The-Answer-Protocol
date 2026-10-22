package component

import (
	"tap-gui/src/graphic/component/ui"
	"tap-gui/src/service"
	"tap-gui/src/utils"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type PlayerComponent struct {
	Manager *service.PlayerManager
	ScrollY float32
}

// NewPlayerComponent creates a player list view backed by mgr.
func NewPlayerComponent(mgr *service.PlayerManager) *PlayerComponent {
	return &PlayerComponent{Manager: mgr}
}

// Render draws the players within rect.
// rect is the component's screen-space bounds.
func (pc *PlayerComponent) Render(rect rl.Rectangle) {
	rl.DrawRectangleRounded(rect, 0.03, 6, utils.ThemeSurface)
	rl.DrawRectangleRoundedLinesEx(rect, 0.03, 6, 2, utils.ThemeBorder)

	title := "Player in Room"
	titleWidth := float32(rl.MeasureText(title, 14))
	rl.DrawText(title, int32(rect.X+(rect.Width-titleWidth)/2), int32(rect.Y+10), 14, utils.ThemeText)
	rl.DrawLineEx(rl.NewVector2(rect.X+10, rect.Y+30), rl.NewVector2(rect.X+rect.Width-10, rect.Y+30), 1, utils.ThemeBorder)

	if pc.Manager == nil || len(pc.Manager.Players) == 0 {
		emptyText := "No players in room"
		tw := float32(rl.MeasureText(emptyText, 11))
		rl.DrawText(emptyText, int32(rect.X+(rect.Width-tw)/2), int32(rect.Y+50), 11, utils.ThemeTextMuted)
		return
	}

	viewport := rl.NewRectangle(rect.X+6, rect.Y+36, rect.Width-12, rect.Height-42)
	contentHeight := float32(len(pc.Manager.Players)) * 32
	ui.UpdateScrollOffset(&pc.ScrollY, viewport, contentHeight)
	rl.BeginScissorMode(int32(viewport.X), int32(viewport.Y), int32(viewport.Width), int32(viewport.Height))
	itemY := viewport.Y + 4 - pc.ScrollY
	rowH := float32(32)

	for _, p := range pc.Manager.Players {
		iconX := rect.X + 16
		iconY := itemY + 4

		rl.DrawCircle(int32(iconX+6), int32(iconY+4), 4, rl.NewColor(24, 160, 178, 255))
		rl.DrawRectangleRounded(rl.NewRectangle(iconX+1, iconY+9, 10, 8), 0.4, 4, rl.NewColor(24, 160, 178, 255))

		rl.DrawText(p.Name, int32(rect.X+40), int32(itemY+8), 11, utils.ThemeText)

		itemY += rowH
	}
	rl.EndScissorMode()
	ui.DrawScrollIndicator(viewport, contentHeight, pc.ScrollY)
}
