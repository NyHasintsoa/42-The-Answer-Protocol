package page

import (
	"tap-gui/src/graphic/component"
	"tap-gui/src/model/enums"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type MenuPage struct {
	BasePage
	SelectedIndex int
	Buttons       []*component.Button
	WindowWidth   int32
	WindowHeight  int32
}

func NewMenuPage(winWidth, winHeight int32) *MenuPage {
	mp := &MenuPage{
		SelectedIndex: 0,
		WindowWidth:   winWidth,
		WindowHeight:  winHeight,
	}
	mp.State = enums.MainMenu

	btnWidth := float32(400)
	btnHeight := float32(50)
	btnX := float32(winWidth-int32(btnWidth)) / 2

	mp.Buttons = []*component.Button{
		component.NewButton(btnX, 500, btnWidth, btnHeight, "Play Game", 24),
		component.NewButton(btnX, 570, btnWidth, btnHeight, "How To Play", 24),
		component.NewButton(btnX, 640, btnWidth, btnHeight, "High Scores", 24),
		component.NewButton(btnX, 710, btnWidth, btnHeight, "Quit Game", 24),
	}

	return mp
}

func (p *MenuPage) EventListener() bool {
	if rl.IsKeyPressed(rl.KeyUp) || rl.IsKeyPressed(rl.KeyW) {
		p.SelectedIndex = (p.SelectedIndex - 1 + len(p.Buttons)) % len(p.Buttons)
	}

	if rl.IsKeyPressed(rl.KeyDown) || rl.IsKeyPressed(rl.KeyS) {
		p.SelectedIndex = (p.SelectedIndex + 1) % len(p.Buttons)
	}

	enterPressed := rl.IsKeyPressed(rl.KeyEnter) || rl.IsKeyPressed(rl.KeyKpEnter)

	if p.Buttons[0].IsClicked || (p.SelectedIndex == 0 && enterPressed) {
		p.NextState = enums.GamePage
	} else if p.Buttons[1].IsClicked || (p.SelectedIndex == 1 && enterPressed) {
		p.NextState = enums.HelpMenu
	} else if p.Buttons[2].IsClicked || (p.SelectedIndex == 2 && enterPressed) {
		p.NextState = enums.HighScoresPage
	} else if p.Buttons[3].IsClicked || (p.SelectedIndex == 3 && enterPressed) {
		p.NextState = enums.QuitPage
	}

	return false
}

func (p *MenuPage) Render() {
	p.EventListener()
	if p.NextState == enums.QuitPage {
		return
	}

	rl.DrawText("PAC-MAN", 320, 180, 72, rl.Yellow)

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
