// Package main starts The Answer Protocol graphical client.
package main

import (
	"fmt"
	"os"

	"tap-gui/src/graphic"
	"tap-gui/src/model"
	"tap-gui/src/service"

	"github.com/joho/godotenv"
)

const (
	// ColorRed is the ANSI escape sequence for red terminal output.
	ColorRed = "\033[31m"
	// ColorYellow is the ANSI escape sequence for yellow terminal output.
	ColorYellow = "\033[33m"
	// ColorReset restores the terminal's default text color.
	ColorReset = "\033[0m"
)

// printCliError writes a categorized error to standard error using the provided color.
// category labels the error; details is the value to print; color is the terminal escape code.
func printCliError(category string, details any, color string) {
	fmt.Fprintf(os.Stderr, "%s[%s]%s %v\n", color, category, ColorReset, details)
}

// run initializes the client window and runs its event loop.
func run() error {
	window := graphic.NewMainWindow(1500, 920, model.GameConfig{}, "The Answer Protocol - 42")
	defer window.Close()

	window.AddEvent()
	window.LoadPage()
	window.Render()
	return nil
}

// main starts the graphical client and exits with an error status if startup fails.
func main() {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		printCliError("Configuration Error", err, ColorRed)
		os.Exit(1)
	}
	if errEnv := service.EnvChecker(); errEnv != nil {
		printCliError("Configuration Error", errEnv, ColorRed)
		os.Exit(1)
	}
	if err := run(); err != nil {
		printCliError("Unexpected Fatal Exception", err, ColorRed)
		os.Exit(1)
	}
}
