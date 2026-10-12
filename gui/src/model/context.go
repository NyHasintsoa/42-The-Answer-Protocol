package model

import "tap-gui/src/utils"

type GameConfig struct {
}

type GameContext struct {
	Config          GameConfig
	ResourceManager *utils.ResourceManager
	PlayerName      string
}
