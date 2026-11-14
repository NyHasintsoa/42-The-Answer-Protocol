package service

// ConnectedPlayer stores the identifier and display name of a visible player.
type ConnectedPlayer struct {
	ID   string
	Name string
}

// PlayerManager stores the players currently visible in the active room.
type PlayerManager struct {
	Players []ConnectedPlayer
}

// NewPlayerManager creates an empty manager for players reported by the server.
func NewPlayerManager() *PlayerManager {
	return &PlayerManager{Players: []ConnectedPlayer{}}
}

// AddPlayer appends a connected player to manager.
// id is the player's identifier; name is the player's display name.
func (manager *PlayerManager) AddPlayer(id, name string) {
	manager.Players = append(manager.Players, ConnectedPlayer{ID: id, Name: name})
}

// SetPlayers replaces the visible room roster with names returned by LOOK.
func (manager *PlayerManager) SetPlayers(names []string) {
	manager.Players = make([]ConnectedPlayer, 0, len(names))
	for _, name := range names {
		manager.Players = append(manager.Players, ConnectedPlayer{ID: name, Name: name})
	}
}

// RemovePlayer removes the player identified by id and reports whether that player existed.
func (manager *PlayerManager) RemovePlayer(id string) bool {
	for index, player := range manager.Players {
		if player.ID == id {
			manager.Players = append(manager.Players[:index], manager.Players[index+1:]...)
			return true
		}
	}
	return false
}
