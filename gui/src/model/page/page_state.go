package page

type PageState int

const (
	LoadingPage PageState = iota
	MainMenu
	HelpMenu
	GamePage
	PlayerNamePage
	HighScoresPage
	QuitPage
)
