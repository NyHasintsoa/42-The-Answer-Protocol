package graphic

import (
	"tap-gui/src/graphic/page"
	"tap-gui/src/model"
	"tap-gui/src/model/enums"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type MainWindow struct {
	Width        int32
	Height       int32
	Title        string
	CurrentState enums.PageState
	CurrentPage  page.Page
	Windows      map[enums.PageState]page.Page
	Closed       bool
	Context      *model.GameContext
}

func NewMainWindow(width, height int32, config model.GameConfig, title string) *MainWindow {
	mw := &MainWindow{
		Width:  width,
		Height: height,
		Title:  title,
		Closed: false,
		Context: &model.GameContext{
			Config: config,
		},
		Windows: make(map[enums.PageState]page.Page),
	}

	rl.SetTraceLogLevel(rl.LogNone)
	rl.InitWindow(mw.Width, mw.Height, mw.Title)
	rl.SetTargetFPS(60)

	return mw
}

func (mw *MainWindow) AddEvent() {
	rl.SetExitKey(0)
}

func (mw *MainWindow) Close() {
	if mw.Closed {
		return
	}
	for _, p := range mw.Windows {
		p.Unload()
	}
	rl.CloseWindow()
	mw.Closed = true
}

func (mw *MainWindow) LoadPage() {
	mw.Windows = map[enums.PageState]page.Page{
		enums.LoadingPage: page.NewLoadingPage(mw.Context),
		enums.MainMenu:    page.NewMenuPage(mw.Width, mw.Height),
	}
	mw.CurrentState = enums.LoadingPage
	mw.CurrentPage = mw.Windows[mw.CurrentState]
	if mw.CurrentPage != nil {
		mw.CurrentPage.Init(mw.Context)
	}
}

func (mw *MainWindow) Render() {
	for !rl.WindowShouldClose() {
		rl.ClearBackground(rl.Black)
		rl.BeginDrawing()

		if mw.CurrentPage == nil {
			rl.EndDrawing()
			break
		}

		mw.CurrentPage.Render()

		if mw.CurrentPage.GetNextState() == enums.QuitPage {
			rl.EndDrawing()
			break
		}

		if mw.CurrentPage.GetNextState() != mw.CurrentState {
			mw.Context = mw.CurrentPage.GetContext()
			mw.CurrentState = mw.CurrentPage.GetNextState()
			mw.CurrentPage = mw.Windows[mw.CurrentState]
			if mw.CurrentPage != nil {
				mw.CurrentPage.SetNextState(mw.CurrentState)
				mw.CurrentPage.Init(mw.Context)
			}
		}
		rl.EndDrawing()
	}
	mw.Close()
}
