package page

import (
	"fmt"

	"tap-gui/src/model"
	pagemodel "tap-gui/src/model/page"
	"tap-gui/src/utils"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type LoadingPage struct {
	BasePage
	LoadingMessage string
	LoadError      string
}

// NewLoadingPage creates a loading page associated with ctx.
func NewLoadingPage(ctx *model.GameContext) *LoadingPage {
	lp := &LoadingPage{
		LoadingMessage: "LOADING RESOURCES ...",
	}
	lp.State = pagemodel.LoadingPage
	lp.NextState = pagemodel.LoadingPage
	lp.Context = ctx
	return lp
}

// SetLoadError stores message for display on the loading screen.
func (p *LoadingPage) SetLoadError(message string) {
	p.LoadError = message
}

// Draw renders the loading progress and any error message in the given window dimensions.
// winWidth and winHeight are the window width and height in pixels.
func (p *LoadingPage) Draw(winWidth, winHeight int32) {
	centerX := winWidth / 2
	centerY := winHeight / 2

	totalBlockHeight := int32(22 + 16 + 24 + 18 + 12)
	startY := centerY - (totalBlockHeight / 2)

	message := p.LoadingMessage
	if p.Context != nil && p.Context.ResourceManager != nil {
		loaded, total := p.Context.ResourceManager.PreloadProgress()
		message = fmt.Sprintf("LOADING RESOURCES  %d / %d", loaded, total)
	}
	txtSize := rl.MeasureText(message, 22)
	rl.DrawText(message, centerX-(txtSize/2), startY, 22, utils.ThemeAccent)

	barWidth := int32(320)
	barX := centerX - barWidth/2
	barY := startY + 40
	rl.DrawRectangle(barX, barY, barWidth, 10, rl.DarkGray)
	if p.Context != nil && p.Context.ResourceManager != nil {
		loaded, total := p.Context.ResourceManager.PreloadProgress()
		if total > 0 {
			progressWidth := int32(float64(barWidth) * float64(loaded) / float64(total))
			rl.DrawRectangle(barX, barY, progressWidth, 10, utils.ThemeAccent)
		}
	}

	if p.LoadError != "" {
		errorSize := rl.MeasureText(p.LoadError, 16)
		rl.DrawText(p.LoadError, centerX-(errorSize/2), barY+24, 16, rl.Red)
	}
}

// Render draws the loading page using the current screen dimensions.
func (p *LoadingPage) Render() {
	p.Draw(int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()))
}
