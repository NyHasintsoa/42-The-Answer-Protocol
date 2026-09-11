package page

import (
	"tap-gui/src/model"
	"tap-gui/src/model/enums"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type LoadingPage struct {
	BasePage
	RotationAngle    float32
	IsGenerationDone bool
	LoadingMessage   string
	ProgressText     string
}

func NewLoadingPage(ctx *model.GameContext) *LoadingPage {
	lp := &LoadingPage{
		RotationAngle:    0.0,
		IsGenerationDone: false,
		LoadingMessage:   "GENERATING LEVELS...",
		ProgressText:     "Level generation progressing",
	}
	lp.State = enums.LoadingPage
	lp.NextState = enums.LoadingPage
	lp.Context = ctx

	go lp.performHeavyGeneration()

	return lp
}

func (p *LoadingPage) performHeavyGeneration() {
	// Simulate background engine warm-up so the loading animation plays on startup
	time.Sleep(2 * time.Second)
	p.IsGenerationDone = true
}

func (p *LoadingPage) Update() {
	p.RotationAngle += 180.0 * rl.GetFrameTime()
	if p.RotationAngle >= 360.0 {
		p.RotationAngle -= 360.0
	}

	if p.IsGenerationDone {
		p.NextState = enums.MainMenu
	}
}

func (p *LoadingPage) Draw(winWidth, winHeight int32) {
	centerX := winWidth / 2
	centerY := winHeight / 2

	rl.DrawCircleSectorLines(
		rl.NewVector2(float32(centerX), float32(centerY-20)),
		45.0,
		p.RotationAngle,
		p.RotationAngle+270.0,
		36,
		rl.Yellow,
	)

	txtSize := rl.MeasureText(p.LoadingMessage, 22)
	rl.DrawText(p.LoadingMessage, centerX-(txtSize/2), centerY+65, 22, rl.Gold)

	subTxt := "Please wait while system matrices align"
	subSize := rl.MeasureText(subTxt, 14)
	rl.DrawText(subTxt, centerX-(subSize/2), centerY+105, 14, rl.DarkGray)
}

func (p *LoadingPage) Render() {
	p.Update()
	p.Draw(1177, 920) // Use your main window dimensions if needed
}