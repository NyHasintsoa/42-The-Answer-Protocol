// Package page defines the states used to navigate between game screens.
package page

// PageState identifies the screen currently shown by the graphical client.
type PageState int

const (
	// LoadingPage identifies the asset-loading screen.
	LoadingPage PageState = iota
	// MainMenu identifies the main menu screen.
	MainMenu
	// GamePage identifies the active gameplay screen.
	GamePage
	// PlayerNamePage identifies the player-name entry screen.
	PlayerNamePage
	// QuitPage requests termination of the graphical client.
	QuitPage
)
