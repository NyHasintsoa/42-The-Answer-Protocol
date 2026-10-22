package service

type ConnectedPlayer struct {
	ID   string
	Name string
}

type PlayerManager struct {
	Players []ConnectedPlayer
}

func NewPlayerManager() *PlayerManager {
	return &PlayerManager{
		Players: []ConnectedPlayer{
			{ID: "p1", Name: "PlayerName 1"},
			{ID: "p2", Name: "PlayerName 2"},
			{ID: "p3", Name: "PlayerName 3"},
		},
	}
}

func (manager *PlayerManager) AddPlayer(id, name string) {
	manager.Players = append(manager.Players, ConnectedPlayer{ID: id, Name: name})
}

func (manager *PlayerManager) RemovePlayer(id string) bool {
	for index, player := range manager.Players {
		if player.ID == id {
			manager.Players = append(manager.Players[:index], manager.Players[index+1:]...)
			return true
		}
	}
	return false
}
