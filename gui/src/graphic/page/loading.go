package page

import (
	"time"

	"tap-gui/src/model"
	"tap-gui/src/model/enums"
	"tap-gui/src/utils"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type LoadingPage struct {
	BasePage
	LoadingMessage string
}

func NewLoadingPage(ctx *model.GameContext) *LoadingPage {
	lp := &LoadingPage{
		LoadingMessage: "LOADING RESOURCES ...",
	}
	lp.State = enums.LoadingPage
	lp.NextState = enums.LoadingPage
	lp.Context = ctx

	go lp.performHeavyGeneration()

	return lp
}

func (p *LoadingPage) performHeavyGeneration() {
	time.Sleep(2 * time.Second)
}

func (p *LoadingPage) Draw(winWidth, winHeight int32) {
	centerX := winWidth / 2
	centerY := winHeight / 2

	totalBlockHeight := int32(22 + 16 + 24)
	startY := centerY - (totalBlockHeight / 2)

	txtSize := rl.MeasureText(p.LoadingMessage, 22)
	rl.DrawText(p.LoadingMessage, centerX-(txtSize/2), startY, 22, utils.ThemeAccent)
}

func (p *LoadingPage) Render() {
	p.Draw(int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()))
}
