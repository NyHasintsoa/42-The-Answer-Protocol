package component

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type ConnectedPlayer struct {
	ID   string
	Name string
}

type PlayerManager struct {
	Players []ConnectedPlayer
}

func NewPlayerManager() *PlayerManager {
	return &PlayerManager{
		Players: []ConnectedPlayer{
			{ID: "p1", Name: "PlayerName 1"},
			{ID: "p2", Name: "PlayerName 2"},
			{ID: "p3", Name: "PlayerName 3"},
		},
	}
}

func (pm *PlayerManager) AddPlayer(id, name string) {
	pm.Players = append(pm.Players, ConnectedPlayer{ID: id, Name: name})
	fmt.Printf("[PLAYER MANAGER] Connected player added: %s (%s)\n", name, id)
}

type PlayerComponent struct {
	Manager *PlayerManager
}

func NewPlayerComponent(mgr *PlayerManager) *PlayerComponent {
	return &PlayerComponent{Manager: mgr}
}

func (pc *PlayerComponent) Render(rect rl.Rectangle) {
	rl.DrawRectangleRounded(rect, 0.03, 6, rl.NewColor(245, 247, 250, 255))
	rl.DrawRectangleRoundedLinesEx(rect, 0.03, 6, 2, rl.NewColor(180, 190, 200, 255))

	title := "Player in Room"
	titleWidth := float32(rl.MeasureText(title, 14))
	rl.DrawText(title, int32(rect.X+(rect.Width-titleWidth)/2), int32(rect.Y+10), 14, rl.NewColor(40, 50, 60, 255))
	rl.DrawLineEx(rl.NewVector2(rect.X+10, rect.Y+30), rl.NewVector2(rect.X+rect.Width-10, rect.Y+30), 1, rl.NewColor(210, 220, 230, 255))

	if pc.Manager == nil || len(pc.Manager.Players) == 0 {
		emptyText := "No players in room"
		tw := float32(rl.MeasureText(emptyText, 11))
		rl.DrawText(emptyText, int32(rect.X+(rect.Width-tw)/2), int32(rect.Y+50), 11, rl.Gray)
		return
	}

	itemY := rect.Y + 40
	rowH := float32(32)

	for _, p := range pc.Manager.Players {
		if itemY+rowH > rect.Y+rect.Height-5 {
			break
		}

		iconX := rect.X + 16
		iconY := itemY + 4

		rl.DrawCircle(int32(iconX+6), int32(iconY+4), 4, rl.NewColor(24, 160, 178, 255))
		rl.DrawRectangleRounded(rl.NewRectangle(iconX+1, iconY+9, 10, 8), 0.4, 4, rl.NewColor(24, 160, 178, 255))

		rl.DrawText(p.Name, int32(rect.X+40), int32(itemY+8), 11, rl.NewColor(40, 50, 60, 255))

		itemY += rowH
	}
}