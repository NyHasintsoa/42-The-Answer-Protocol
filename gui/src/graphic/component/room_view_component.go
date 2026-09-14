package component

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type ExitDirection string

const (
	DirNorth ExitDirection = "NORTH"
	DirSouth ExitDirection = "SOUTH"
	DirEast  ExitDirection = "EAST"
	DirWest  ExitDirection = "WEST"
)

type RoomViewData struct {
	Title       string
	Description string
	Exits       map[ExitDirection]bool
}

type RoomViewManager struct {
	CurrentRoom RoomViewData
}

func NewRoomViewManager() *RoomViewManager {
	return &RoomViewManager{
		CurrentRoom: RoomViewData{
			Title:       "Village Square",
			Description: "Praesent commodo cursus magna, vel scelerisque nisl consectetur et. Vivamus sagittis lacus vel augue laoreet rutrum faucibus dolor auctor.\nPraesent commodo cursus magna, vel scelerisque nisl consectetur et. Vivamus sagittis lacus vel augue laoreet rutrum faucibus dolor auctor.",
			Exits: map[ExitDirection]bool{
				DirNorth: true,
				DirSouth: true,
				DirEast:  true,
				DirWest:  true,
			},
		},
	}
}

func (rvm *RoomViewManager) Move(dir ExitDirection) {
	fmt.Printf("[ROOM ACTION] Moved direction: %s\n", dir)
}

type RoomViewComponent struct {
	Manager *RoomViewManager
}

func NewRoomViewComponent(mgr *RoomViewManager) *RoomViewComponent {
	return &RoomViewComponent{Manager: mgr}
}

func (rvc *RoomViewComponent) Render(rect rl.Rectangle) {
	rl.DrawRectangleRounded(rect, 0.03, 6, rl.NewColor(245, 247, 250, 255))
	rl.DrawRectangleRoundedLinesEx(rect, 0.03, 6, 2, rl.NewColor(180, 190, 200, 255))

	title := "Room View"
	titleWidth := float32(rl.MeasureText(title, 14))
	rl.DrawText(title, int32(rect.X+(rect.Width-titleWidth)/2), int32(rect.Y+10), 14, rl.NewColor(40, 50, 60, 255))
	rl.DrawLineEx(rl.NewVector2(rect.X+10, rect.Y+30), rl.NewVector2(rect.X+rect.Width-10, rect.Y+30), 1, rl.NewColor(210, 220, 230, 255))

	if rvc.Manager != nil && rvc.Manager.CurrentRoom.Description != "" {
		drawTextWrapped(
			rvc.Manager.CurrentRoom.Description,
			int32(rect.X+12),
			int32(rect.Y+36),
			int32(rect.Width-24),
			10,
			rl.NewColor(60, 70, 80, 255),
		)
	}

	directions := []struct {
		Label string
		Dir   ExitDirection
	}{
		{"NORD", DirNorth},
		{"SOUTH", DirSouth},
		{"EAST", DirEast},
		{"WEST", DirWest},
	}

	btnCount := float32(len(directions))
	btnMargin := float32(6)
	totalSpacing := btnMargin * (btnCount + 1)
	btnW := (rect.Width - totalSpacing) / btnCount
	btnH := float32(26)
	btnY := rect.Y + rect.Height - btnH - 8

	mousePos := rl.GetMousePosition()
	clicked := rl.IsMouseButtonPressed(rl.MouseLeftButton)

	tealBtnColor := rl.NewColor(24, 160, 178, 255)
	tealBtnHover := rl.NewColor(18, 130, 146, 255)

	for i, d := range directions {
		btnX := rect.X + btnMargin + float32(i)*(btnW+btnMargin)
		btnRect := rl.NewRectangle(btnX, btnY, btnW, btnH)

		color := tealBtnColor
		if rl.CheckCollisionPointRec(mousePos, btnRect) {
			color = tealBtnHover
			if clicked {
				fmt.Printf("[ACTION LOG] Movement button clicked: %s\n", d.Label)
				if rvc.Manager != nil {
					rvc.Manager.Move(d.Dir)
				}
			}
		}

		rl.DrawRectangleRounded(btnRect, 0.2, 4, color)
		textW := float32(rl.MeasureText(d.Label, 10))
		rl.DrawText(d.Label, int32(btnX+(btnW-textW)/2), int32(btnY+7), 10, rl.White)
	}
}
