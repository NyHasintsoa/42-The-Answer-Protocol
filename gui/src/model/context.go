// Package model defines shared state passed between game pages and services.
package model

import (
	"tap-gui/src/service"
	"tap-gui/src/utils"
)

// GameConfig contains configuration shared by the graphical client.
type GameConfig struct {
	MapJSON string
}

// GameContext holds state shared across the client's pages.
type GameContext struct {
	Config          GameConfig
	ResourceManager *utils.ResourceManager
	PlayerName      string
	TCPClient       *service.TCPClient
}
