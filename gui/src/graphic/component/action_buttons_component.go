package component

import (
	"fmt"
	"tap-gui/src/service"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type ActionButtonsComponent struct {
	InventoryMgr *service.InventoryManager
	ChatMgr      *service.ChatManager
	QuestMgr     *service.QuestManager
	CommandInput *InputComponent
}

func NewActionButtonsComponent(inv *service.InventoryManager, chat *service.ChatManager, quest *service.QuestManager) *ActionButtonsComponent {
	return &ActionButtonsComponent{
		InventoryMgr: inv,
		ChatMgr:      chat,
		QuestMgr:     quest,
		CommandInput: NewInputComponent(0, 0, 100, 32, "Command Here ...", 100),
	}
}

func (abc *ActionButtonsComponent) Render(rect rl.Rectangle) {
	mousePos := rl.GetMousePosition()
	clicked := rl.IsMouseButtonPressed(rl.MouseLeftButton)

	inputH := float32(32)
	gridH := rect.Height - inputH - 8

	buttons := [][]struct {
		Label string
		IsRed bool
	}{
		{
			{"INVENTORY", false},
			{"QUESTS", false},
			{"STATUS", false},
			{"GROUP", false},
		},
		{
			{"CHAT", false},
			{"WHO", false},
			{"LOOK", false},
			{"QUIT", true},
		},
	}

	rowHeight := (gridH - 6) / 2
	blueBg := rl.NewColor(0, 140, 215, 255)
	blueHover := rl.NewColor(0, 170, 240, 255)
	redBg := rl.NewColor(220, 40, 50, 255)
	redHover := rl.NewColor(240, 70, 80, 255)

	for r, row := range buttons {
		colWidth := (rect.Width - float32(len(row)-1)*6) / float32(len(row))
		for c, btn := range row {
			bx := rect.X + float32(c)*(colWidth+6)
			by := rect.Y + float32(r)*(rowHeight+6)
			btnRect := rl.NewRectangle(bx, by, colWidth, rowHeight)

			bg := blueBg
			hover := blueHover
			if btn.IsRed {
				bg = redBg
				hover = redHover
			}

			currentColor := bg
			if rl.CheckCollisionPointRec(mousePos, btnRect) {
				currentColor = hover
				if clicked {
					fmt.Printf("[ACTION LOG] Bottom button clicked: %s\n", btn.Label)
					abc.handleAction(btn.Label)
				}
			}

			rl.DrawRectangleRounded(btnRect, 0.15, 6, currentColor)
			rl.DrawRectangleRoundedLinesEx(btnRect, 0.15, 6, 2, rl.NewColor(255, 255, 255, 100))

			fontSize := int32(11)
			tw := float32(rl.MeasureText(btn.Label, fontSize))
			ty := by + (rowHeight-float32(fontSize))/2
			rl.DrawText(btn.Label, int32(bx+(colWidth-tw)/2), int32(ty), fontSize, rl.White)
		}
	}

	inputY := rect.Y + gridH + 8
	abc.CommandInput.Rect = rl.NewRectangle(rect.X, inputY, rect.Width, inputH)

	dt := rl.GetFrameTime()
	if abc.CommandInput.Update(dt) {
		cmd := abc.CommandInput.GetAndClearText()
		if cmd != "" {
			fmt.Printf("[COMMAND LINE LOG] Executing command: %s\n", cmd)
		}
	}
	abc.CommandInput.Render()
}

func (abc *ActionButtonsComponent) handleAction(label string) {
	switch label {
	case "INVENTORY":
		if abc.InventoryMgr != nil {
			abc.InventoryMgr.Toggle()
		}
	case "CHAT":
		if abc.ChatMgr != nil {
			abc.ChatMgr.Toggle()
		}
	case "QUESTS":
		if abc.QuestMgr != nil {
			abc.QuestMgr.Toggle()
		}
	case "STATUS", "GROUP", "LOOK", "WHO", "QUIT":
	}
}