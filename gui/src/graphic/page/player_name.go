package page

import (
	"context"
	"strings"
	"sync/atomic"
	"time"

	"tap-gui/src/graphic/component/ui"
	"tap-gui/src/model"
	pagemodel "tap-gui/src/model/page"
	"tap-gui/src/utils"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type PlayerNamePage struct {
	BasePage
	name_input     ui.InputComponent
	submit_btn     ui.Button
	winWidth       int32
	winHeight      int32
	errorMessage   string
	connecting     bool
	playerName     string
	connectCancel  context.CancelFunc
	connectResults chan connectionResult
	unloaded       atomic.Bool
}

type connectionResult struct {
	err error
}

// NewPlayerNamePage creates the player-name form for the specified window and game context.
// winWidth and winHeight are the window dimensions in pixels; ctx provides shared game state.
func NewPlayerNamePage(
	winWidth, winHeight int32, ctx *model.GameContext,
) *PlayerNamePage {
	mainPage := &PlayerNamePage{
		winWidth:       winWidth,
		winHeight:      winHeight,
		connectResults: make(chan connectionResult, 1),
	}
	mainPage.Context = ctx
	mainPage.State = pagemodel.PlayerNamePage
	mainPage.NextState = pagemodel.PlayerNamePage

	cardW := float32(520)
	cardH := float32(380)
	cardX := (float32(winWidth) - cardW) / 2
	cardY := (float32(winHeight) - cardH) / 2

	inputW := float32(400)
	inputH := float32(50)
	inputX := cardX + (cardW-inputW)/2
	inputY := cardY + 80

	btnW := float32(220)
	btnH := float32(50)
	btnX := cardX + (cardW-btnW)/2
	btnY := cardY + cardH - 75

	mainPage.name_input = *ui.NewInputComponent(inputX, inputY, inputW, inputH, "Enter Player Name", 20)

	btn := ui.NewButton(btnX, btnY, btnW, btnH, "CONTINUE", 18)
	btn.BorderWidth = 2
	btn.BorderRadius = 0.25
	mainPage.submit_btn = *btn

	return mainPage
}

// handleSubmit validates the entered name and starts a server connection attempt.
func (p *PlayerNamePage) handleSubmit() {
	if p.connecting {
		return
	}
	rawName := strings.TrimSpace(p.name_input.Text)

	if rawName == "" {
		p.errorMessage = "PLAYER NAME CANNOT BE EMPTY."
		return
	}
	if strings.ToLower(rawName) == "shadowslayer" {
		p.errorMessage = "PLAYER NAME IS ALREADY IN USE."
		return
	}
	if p.Context == nil || p.Context.TCPClient == nil {
		p.errorMessage = "SERVER CAN'T BE REACHED."
		return
	}

	p.playerName = rawName
	p.errorMessage = ""
	p.connecting = true
	p.submit_btn.Text = "CONNECTING..."
	connectContext, cancel := context.WithCancel(context.Background())
	p.connectCancel = cancel
	client := p.Context.TCPClient

	go func() {
		err := client.Connect(connectContext, 3*time.Second)
		if !p.unloaded.Load() {
			p.connectResults <- connectionResult{err: err}
		}
	}()
}

// Update handles player-name input and submission.
func (p *PlayerNamePage) Update() {
	dt := rl.GetFrameTime()
	select {
	case result := <-p.connectResults:
		p.connecting = false
		if p.connectCancel != nil {
			p.connectCancel()
			p.connectCancel = nil
		}
		p.submit_btn.Text = "CONTINUE"
		if result.err != nil {
			p.errorMessage = "SERVER CAN'T BE REACHED."
			return
		}
		if result.err != nil {
			p.errorMessage = "SERVER CAN'T BE REACHED."
			return
		}
		p.Context.PlayerName = p.playerName
		p.NextState = pagemodel.GamePage
	default:
	}

	if p.connecting {
		return
	}

	submitted := p.name_input.Update(dt)

	if p.submit_btn.IsClicked || submitted || rl.IsKeyPressed(rl.KeyEnter) {
		p.handleSubmit()
	}
}

// Unload releases resources owned by the player-name page.
func (p *PlayerNamePage) Unload() {
	p.unloaded.Store(true)
	if p.connectCancel != nil {
		p.connectCancel()
		p.connectCancel = nil
	}
	p.BasePage.Unload()
}

// Render draws the player-name entry page.
func (p *PlayerNamePage) Render() {
	p.Update()

	rl.ClearBackground(utils.ThemeBackground)
	rl.DrawRectangle(0, 0, p.winWidth, p.winHeight, rl.NewColor(utils.ThemeBackground.R, utils.ThemeBackground.G, utils.ThemeBackground.B, 220))

	cardW := float32(520)
	cardH := float32(380)
	cardX := (float32(p.winWidth) - cardW) / 2
	cardY := (float32(p.winHeight) - cardH) / 2
	cardRect := rl.NewRectangle(cardX, cardY, cardW, cardH)

	rl.DrawRectangleRounded(cardRect, 0.05, 12, utils.ThemeSurface)
	rl.DrawRectangleRoundedLinesEx(cardRect, 0.05, 12, 3, utils.ThemeBorderStrong)

	titleText := "Player name"
	titleFont := int32(28)
	titleW := rl.MeasureText(titleText, titleFont)
	rl.DrawText(titleText, int32(cardX+(cardW-float32(titleW))/2), int32(cardY+28), titleFont, utils.ThemeText)

	p.name_input.Render()

	if p.errorMessage != "" {
		errFont := int32(14)
		errW := rl.MeasureText(p.errorMessage, errFont)
		errX := int32(cardX + (cardW-float32(errW))/2)
		errY := int32(cardY + 220)

		rl.DrawText(p.errorMessage, errX+1, errY+1, errFont, rl.NewColor(80, 10, 20, 255))
		rl.DrawText(p.errorMessage, errX, errY, errFont, utils.ThemeDanger)
	}
	p.submit_btn.Render()
}
