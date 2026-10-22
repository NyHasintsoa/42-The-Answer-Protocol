package page

import (
	"tap-gui/src/graphic/component/ui"
	pagemodel "tap-gui/src/model/page"
	"tap-gui/src/utils"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type MenuPage struct {
	BasePage
	SelectedIndex int
	Buttons       []*ui.Button
	WindowWidth   int32
	WindowHeight  int32
}

// NewMenuPage creates the main menu for a window with the specified dimensions.
// winWidth and winHeight are the window width and height in pixels.
func NewMenuPage(winWidth int32, winHeight int32) *MenuPage {
	mp := &MenuPage{
		SelectedIndex: 0,
		WindowWidth:   winWidth,
		WindowHeight:  winHeight,
	}
	mp.State = pagemodel.MainMenu

	btnWidth := float32(400)
	btnHeight := float32(50)
	btnX := float32(winWidth-int32(btnWidth)) / 2

	mp.Buttons = []*ui.Button{
		ui.NewButton(btnX, 500, btnWidth, btnHeight, "Play Game", 24),
		ui.NewButton(btnX, 710, btnWidth, btnHeight, "Quit Game", 24),
	}

	return mp
}

// EventListener processes keyboard and button input for the menu.
func (p *MenuPage) EventListener() {
	if rl.IsKeyPressed(rl.KeyUp) || rl.IsKeyPressed(rl.KeyW) {
		p.SelectedIndex = (p.SelectedIndex - 1 + len(p.Buttons)) % len(p.Buttons)
	}

	if rl.IsKeyPressed(rl.KeyDown) || rl.IsKeyPressed(rl.KeyS) {
		p.SelectedIndex = (p.SelectedIndex + 1) % len(p.Buttons)
	}

	enterPressed := rl.IsKeyPressed(rl.KeyEnter) || rl.IsKeyPressed(rl.KeyKpEnter)

	if p.Buttons[0].IsClicked || (p.SelectedIndex == 0 && enterPressed) {
		p.NextState = pagemodel.PlayerNamePage
	} else if p.Buttons[1].IsClicked || (p.SelectedIndex == 3 && enterPressed) {
		p.NextState = pagemodel.QuitPage
	}
}

// Render draws the menu and updates its selected page state.
func (p *MenuPage) Render() {
	p.EventListener()
	if p.NextState == pagemodel.QuitPage {
		return
	}

	rl.DrawText("The Answer Protocol", 320, 180, 72, utils.ThemeAccent)

	mousePos := rl.GetMousePosition()
	for i, button := range p.Buttons {
		if rl.CheckCollisionPointRec(mousePos, button.Rect) {
			p.SelectedIndex = i
		}
	}

	for _, button := range p.Buttons {
		button.Render()
	}
}
