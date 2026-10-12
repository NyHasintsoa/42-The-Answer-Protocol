package page

type PageState int

const (
	LoadingPage PageState = iota
	MainMenu
	GamePage
	PlayerNamePage
	QuitPage
)
