package main

import (
	"fmt"
	"os"

	"tap-gui/src/graphic"
	"tap-gui/src/model"
)

const (
	ColorRed    = "\033[31m"
	ColorYellow = "\033[33m"
	ColorReset  = "\033[0m"
)

func printCliError(category string, details interface{}, color string) {
	fmt.Fprintf(os.Stderr, "%s[%s]%s %v\n", color, category, ColorReset, details)
}

func run(args []string) error {
	configFile := "config.json"
	if len(args) > 0 {
		configFile = args[0]
	}
	_ = configFile

	config := model.GameConfig{Lives: 3}
	window := graphic.NewMainWindow(1500, 1000, config, "Pac-Man")
	defer window.Close()

	window.AddEvent()
	window.LoadPage()
	window.Render()
	return nil
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		printCliError("Unexpected Fatal Exception", err, ColorRed)
		os.Exit(1)
	}
}