// Package graphic manages the Raylib window and transitions between game pages.
package graphic

import (
	"tap-gui/src/graphic/page"
	"tap-gui/src/model"
	pagemodel "tap-gui/src/model/page"
	"tap-gui/src/service"
	"tap-gui/src/utils"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type MainWindow struct {
	Width        int32
	Height       int32
	Title        string
	CurrentState pagemodel.PageState
	CurrentPage  page.Page
	Windows      map[pagemodel.PageState]page.Page
	Closed       bool
	Context      *model.GameContext
	pagesLoaded  bool
}

// NewMainWindow creates and initializes the game's Raylib window.
// width and height are the window dimensions in pixels; config is the game configuration;
// title is the window title shown by the operating system.
func NewMainWindow(width, height int32, config model.GameConfig, title string) *MainWindow {
	mw := &MainWindow{
		Width:  width,
		Height: height,
		Title:  title,
		Closed: false,
		Context: &model.GameContext{
			Config:          config,
			ResourceManager: utils.NewResourceManager(),
			TCPClient:       service.NewTCPClient(),
		},
		Windows: make(map[pagemodel.PageState]page.Page),
	}
	rl.SetTraceLogLevel(rl.LogNone)
	rl.InitWindow(mw.Width, mw.Height, mw.Title)
	rl.SetTargetFPS(60)

	return mw
}

// AddEvent disables Raylib's default exit-key handling.
func (mw *MainWindow) AddEvent() {
	rl.SetExitKey(0)
}

// Close unloads pages and shared resources, closes the network client, and shuts down the window.
func (mw *MainWindow) Close() {
	if mw.Closed {
		return
	}
	for _, p := range mw.Windows {
		if p != nil {
			p.Unload()
		}
	}
	mw.CurrentPage = nil
	mw.Windows = nil
	if mw.Context != nil && mw.Context.ResourceManager != nil {
		mw.Context.ResourceManager.UnloadAll()
	}
	if mw.Context != nil && mw.Context.TCPClient != nil {
		_ = mw.Context.TCPClient.Close()
	}
	rl.CloseWindow()
	mw.Closed = true
}

// LoadPage initializes the loading page and sets it as the current page.
func (mw *MainWindow) LoadPage() {
	mw.Windows = map[pagemodel.PageState]page.Page{
		pagemodel.LoadingPage: page.NewLoadingPage(mw.Context),
	}
	mw.pagesLoaded = false
	mw.CurrentState = pagemodel.LoadingPage
	mw.CurrentPage = mw.Windows[mw.CurrentState]
	if mw.CurrentPage != nil {
		mw.CurrentPage.Init(mw.Context)
	}
}

// loadPages creates the player-name, menu, and gameplay pages.
func (mw *MainWindow) loadPages() {
	mw.Windows[pagemodel.PlayerNamePage] = page.NewPlayerNamePage(mw.Width, mw.Height, mw.Context)
	mw.Windows[pagemodel.MainMenu] = page.NewMenuPage(mw.Width, mw.Height)
	mw.Windows[pagemodel.GamePage] = page.NewGamePage(mw.Width, mw.Height, mw.Context)
}

// Render runs the window loop, draws the active page, and processes page transitions.
func (mw *MainWindow) Render() {
	for !rl.WindowShouldClose() {
		rl.ClearBackground(utils.ThemeBackground)
		rl.BeginDrawing()

		if mw.CurrentPage == nil {
			rl.EndDrawing()
			break
		}

		mw.CurrentPage.Render()

		if mw.CurrentPage.GetNextState() == pagemodel.QuitPage {
			rl.EndDrawing()
			break
		}

		if mw.CurrentPage.GetNextState() != mw.CurrentState {
			nextState := mw.CurrentPage.GetNextState()
			mw.Context = mw.CurrentPage.GetContext()
			mw.CurrentState = nextState
			mw.CurrentPage = mw.Windows[mw.CurrentState]
			if mw.CurrentPage != nil {
				mw.CurrentPage.SetNextState(mw.CurrentState)
				mw.CurrentPage.Init(mw.Context)
			}
		}
		rl.EndDrawing()

		if loadingPage, ok := mw.CurrentPage.(*page.LoadingPage); ok && !mw.pagesLoaded && loadingPage.LoadError == "" {
			if mw.Context == nil || mw.Context.ResourceManager == nil {
				loadingPage.SetLoadError("resource manager is not initialized")
			} else {
				done, err := mw.Context.ResourceManager.PreloadNextBatch("assets", 8)
				if err != nil {
					loadingPage.SetLoadError(err.Error())
				} else if done {
					mw.loadPages()
					mw.pagesLoaded = true
					loadingPage.SetNextState(pagemodel.MainMenu)
				}
			}
		}
	}
	mw.Close()
}
