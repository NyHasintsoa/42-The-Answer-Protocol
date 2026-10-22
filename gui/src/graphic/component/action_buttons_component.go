// Package component provides the visual and interactive pieces of game pages.
package component

import (
	"fmt"
	"strings"
	"tap-gui/src/graphic/component/ui"
	"tap-gui/src/service"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type ActionButtonsComponent struct {
	InventoryMgr *service.InventoryManager
	ChatMgr      *service.ChatManager
	QuestMgr     *service.QuestManager
	CommandInput *ui.InputComponent
	OnLog        func(string)
	OnQuit       func()
}

// NewActionButtonsComponent creates action buttons backed by the supplied managers.
// inv controls inventory actions; chat controls chat actions; quest controls quest actions.
func NewActionButtonsComponent(inv *service.InventoryManager, chat *service.ChatManager, quest *service.QuestManager) *ActionButtonsComponent {
	return &ActionButtonsComponent{
		InventoryMgr: inv,
		ChatMgr:      chat,
		QuestMgr:     quest,
		CommandInput: ui.NewInputComponent(0, 0, 100, 20, "Command Here ...", 100),
	}
}

// Render draws the action buttons and command input within rect.
// rect is the component's screen-space bounds.
func (abc *ActionButtonsComponent) Render(rect rl.Rectangle) {
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

	for r, row := range buttons {
		colWidth := (rect.Width - float32(len(row)-1)*6) / float32(len(row))
		for c, btn := range row {
			bx := rect.X + float32(c)*(colWidth+6)
			by := rect.Y + float32(r)*(rowHeight+6)
			button := ui.NewButton(bx, by, colWidth, rowHeight, btn.Label, 11)
			button.BorderWidth = 2
			button.BorderRadius = 0.15
			button.TextColor = rl.White
			button.ShadowColor = rl.Blank
			if btn.IsRed {
				button.BgColor = rl.NewColor(220, 40, 50, 255)
				button.HoverBgColor = rl.NewColor(240, 70, 80, 255)
			} else {
				button.BgColor = rl.NewColor(0, 140, 215, 255)
				button.HoverBgColor = rl.NewColor(0, 170, 240, 255)
			}
			button.ClickedColor = button.HoverBgColor
			button.Render()
			if button.IsClicked {
				if abc.OnLog != nil {
					abc.OnLog(fmt.Sprintf("Selected %s.", strings.ToLower(btn.Label)))
				}
				abc.handleAction(btn.Label)
			}
		}
	}

	inputY := rect.Y + gridH + 8
	abc.CommandInput.Rect = rl.NewRectangle(rect.X, inputY, rect.Width, inputH)

	dt := rl.GetFrameTime()
	if abc.CommandInput.Update(dt) {
		cmd := abc.CommandInput.GetAndClearText()
		if cmd != "" {
			if abc.OnLog != nil {
				abc.OnLog(fmt.Sprintf("Entered command: %s", cmd))
			}
		}
	}
	abc.CommandInput.Render()
}

// handleAction dispatches the action associated with label to its manager or callback.
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
	case "QUIT":
		if abc.OnQuit != nil {
			abc.OnQuit()
		}
	case "STATUS", "GROUP", "LOOK", "WHO":
	}
}
