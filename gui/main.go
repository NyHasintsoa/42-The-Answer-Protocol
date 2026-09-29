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

func run() error {
	config := model.GameConfig{}
	window := graphic.NewMainWindow(1500, 1000, config, "The Answer Protocol - 42")
	defer window.Close()

	window.AddEvent()
	window.LoadPage()
	window.Render()
	return nil
}

func main() {
	if err := run(); err != nil {
		printCliError("Unexpected Fatal Exception", err, ColorRed)
		os.Exit(1)
	}
}
