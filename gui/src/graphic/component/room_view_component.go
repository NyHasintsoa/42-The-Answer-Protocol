package component

import (
	"tap-gui/src/graphic/component/ui"
	"tap-gui/src/utils"

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
	OnMove      func(ExitDirection)
	IsMoving    bool
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
	if !rvm.CurrentRoom.Exits[dir] {
		return
	}
	if rvm.OnMove != nil {
		rvm.OnMove(dir)
	}
}

type RoomViewComponent struct {
	Manager *RoomViewManager
	ScrollY float32
}

func NewRoomViewComponent(mgr *RoomViewManager) *RoomViewComponent {
	return &RoomViewComponent{Manager: mgr}
}

func (rvc *RoomViewComponent) Render(rect rl.Rectangle) {
	rl.DrawRectangleRounded(rect, 0.03, 6, utils.ThemeSurface)
	rl.DrawRectangleRoundedLinesEx(rect, 0.03, 6, 2, utils.ThemeBorder)

	title := "Room View"
	if rvc.Manager != nil && rvc.Manager.CurrentRoom.Title != "" {
		title = rvc.Manager.CurrentRoom.Title
	}
	titleWidth := float32(rl.MeasureText(title, 14))
	rl.DrawText(title, int32(rect.X+(rect.Width-titleWidth)/2), int32(rect.Y+10), 14, utils.ThemeText)
	rl.DrawLineEx(rl.NewVector2(rect.X+10, rect.Y+30), rl.NewVector2(rect.X+rect.Width-10, rect.Y+30), 1, utils.ThemeBorder)

	directions := []struct {
		Label string
		Dir   ExitDirection
	}{
		{"NORTH", DirNorth},
		{"SOUTH", DirSouth},
		{"EAST", DirEast},
		{"WEST", DirWest},
	}

	btnH := float32(26)
	btnY := rect.Y + rect.Height - btnH - 8
	descriptionViewport := rl.NewRectangle(rect.X+8, rect.Y+36, rect.Width-22, btnY-rect.Y-42)
	var description string
	if rvc.Manager != nil {
		description = rvc.Manager.CurrentRoom.Description
	}
	descriptionLines := ui.WrapTextLines(description, 10, int32(descriptionViewport.Width-8))
	descriptionHeight := float32(len(descriptionLines) * 12)
	ui.UpdateScrollOffset(&rvc.ScrollY, descriptionViewport, descriptionHeight)
	rl.BeginScissorMode(int32(descriptionViewport.X), int32(descriptionViewport.Y), int32(descriptionViewport.Width), int32(descriptionViewport.Height))
	ui.DrawTextWrapped(description, int32(descriptionViewport.X+4), int32(descriptionViewport.Y-rvc.ScrollY), int32(descriptionViewport.Width-8), 10, utils.ThemeTextMuted)
	rl.EndScissorMode()
	ui.DrawScrollIndicator(descriptionViewport, descriptionHeight, rvc.ScrollY)

	btnCount := float32(len(directions))
	btnMargin := float32(6)
	totalSpacing := btnMargin * (btnCount + 1)
	btnW := (rect.Width - totalSpacing) / btnCount
	for i, d := range directions {
		btnX := rect.X + btnMargin + float32(i)*(btnW+btnMargin)
		button := ui.NewButton(btnX, btnY, btnW, btnH, d.Label, 10)
		button.BgColor = rl.NewColor(24, 160, 178, 255)
		button.HoverBgColor = rl.NewColor(18, 130, 146, 255)
		button.ClickedColor = button.HoverBgColor
		button.TextColor = rl.White
		button.BorderColor = rl.Blank
		button.ShadowColor = rl.Blank
		button.BorderWidth = 0
		button.BorderRadius = 0.2
		button.Disabled = rvc.Manager == nil || rvc.Manager.IsMoving || !rvc.Manager.CurrentRoom.Exits[d.Dir]
		button.Render()
		if button.IsClicked {
			if rvc.Manager != nil {
				rvc.Manager.Move(d.Dir)
			}
		}
	}
}
