package main

import (
	"errors"
	"fmt"
	"os"

	"tap-client/src/service"
	"tap-client/src/ui"

	tea "charm.land/bubbletea/v2"
	"github.com/joho/godotenv"
)

const (
	ColorRed    = "\033[31m"
	ColorYellow = "\033[33m"
	ColorReset  = "\033[0m"
)

// PrintError writes a categorized error to standard error using the provided
// color. category labels the error; details is the value to print; color is the
// terminal escape code.
func printError(category string, details any, color string) {
	fmt.Fprintf(os.Stderr, "%s[%s]%s %v\n", color, category, ColorReset, details)
}

// Run boots the TAP terminal client: it loads the environment, validates the
// required configuration, resolves the server address and runs the Bubbletea
// program. Any failure is reported through PrintError and the process exits.
func run() error {
	addr, err := service.ServerAddress()
	if err != nil {
		return err
	}
	program := tea.NewProgram(ui.NewModel(addr, "Loaded environment - server "+addr))
	if _, err := program.Run(); err != nil {
		return err
	}
	return nil
}

func main() {
	if err := godotenv.Load("../.env", ".env"); err != nil && !errors.Is(err, os.ErrNotExist) {
		printError("Configuration Error", fmt.Errorf("load .env: %w", err), ColorRed)
		os.Exit(1)
	}
	if err := service.EnvChecker(); err != nil {
		printError("Configuration Error", err, ColorRed)
		os.Exit(1)
	}
	if err := run(); err != nil {
		printError("Error running TAP CLI client: %v\n", err, ColorRed)
		os.Exit(1)
	}
}
