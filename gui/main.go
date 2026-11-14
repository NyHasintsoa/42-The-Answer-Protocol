// Package main starts The Answer Protocol graphical client.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"tap-gui/src/graphic"
	"tap-gui/src/model"
	"tap-gui/src/service"

	"github.com/joho/godotenv"
)

const (
	ColorRed    = "\033[31m"
	ColorYellow = "\033[33m"
	ColorReset  = "\033[0m"
)

// printCliError writes a categorized error to standard error using the provided color.
// category labels the error; details is the value to print; color is the terminal escape code.
func printCliError(category string, details any, color string) {
	fmt.Fprintf(os.Stderr, "%s[%s]%s %v\n", color, category, ColorReset, details)
}

// run initializes the client window and runs its event loop.
func run() error {
	mapJSON, err := loadServerMap()
	if err != nil {
		return err
	}
	config := model.GameConfig{MapJSON: mapJSON}
	window := graphic.NewMainWindow(1500, 920, config, "The Answer Protocol - 42")
	defer window.Close()

	window.AddEvent()
	window.LoadPage()
	window.Render()
	return nil
}

func loadServerMap() (string, error) {
	configuredPath := os.Getenv("SERVER_MAP_PATH")
	candidates := []string{configuredPath, filepath.Join("..", "server", "map.json"), filepath.Join("server", "map.json")}
	for _, path := range candidates {
		if path == "" {
			continue
		}
		data, err := os.ReadFile(path)
		if err == nil {
			var config service.MapConfig
			if err := json.Unmarshal(data, &config); err != nil {
				return "", fmt.Errorf("decode server map %q: %w", path, err)
			}
			if len(config.World.Rooms) == 0 {
				return "", fmt.Errorf("server map %q has no world rooms", path)
			}
			if _, exists := config.World.Rooms["start"]; !exists {
				return "", fmt.Errorf("server map %q has no start room", path)
			}
			for roomID, room := range config.World.Rooms {
				if room.Name == "" {
					return "", fmt.Errorf("server map %q has an unnamed room %q", path, roomID)
				}
				for _, destination := range room.Exits {
					if _, exists := config.World.Rooms[destination]; !exists {
						return "", fmt.Errorf("server map %q room %q exits to unknown room %q", path, roomID, destination)
					}
				}
			}
			return string(data), nil
		}
		if configuredPath != "" && path == configuredPath {
			return "", fmt.Errorf("read server map %q: %w", path, err)
		}
		if !os.IsNotExist(err) {
			return "", fmt.Errorf("read server map %q: %w", path, err)
		}
	}
	return "", fmt.Errorf("could not find server map.json; set SERVER_MAP_PATH or run the GUI beside ../server/map.json")
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
