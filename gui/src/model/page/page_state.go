// Package page defines the states used to navigate between game screens.
package page

type PageState int

const (
	LoadingPage PageState = iota
	MainMenu
	GamePage
	PlayerNamePage
	QuitPage
)
