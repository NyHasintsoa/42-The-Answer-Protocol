package service

import (
	"fmt"
	"strings"

	"tap-server/src/model"
)

type WorldManager struct {
	world    *WorldService
	presence *BroadcastService
	npcs     *NpcManager
}

func NewWorldManager(world *WorldService, presence *BroadcastService, npcs *NpcManager) *WorldManager {
	return &WorldManager{world: world, presence: presence, npcs: npcs}
}

func (manager *WorldManager) Room(roomID string) (*model.Room, bool) {
	return manager.world.GetRoom(roomID)
}

func (manager *WorldManager) Look(player *model.Player) (*model.LookResponse, bool) {
	room, ok := manager.Room(player.CurrentRoom)
	if !ok {
		return nil, false
	}
	return &model.LookResponse{
		Room: model.RoomSummary{
			ID:          room.ID,
			Name:        room.Name,
			Description: room.Description,
			Exits:       room.Exits,
		},
		Players: manager.PlayersInRoom(player.CurrentRoom),
		Items:   room.Items,
		NPCs:    room.NPCs,
	}, true
}

func (manager *WorldManager) Move(player *model.Player, direction string) (string, bool) {
	room, ok := manager.Room(player.CurrentRoom)
	if !ok {
		return "", false
	}
	nextRoom, ok := room.Exits[strings.ToLower(strings.TrimSpace(direction))]
	if !ok {
		return "", false
	}
	oldRoom := player.CurrentRoom
	manager.Relocate(player, nextRoom)
	return oldRoom, true
}

func (manager *WorldManager) Relocate(player *model.Player, roomID string) {
	oldRoom := player.CurrentRoom
	player.CurrentRoom = roomID
	player.InCombatWith = ""
	manager.BroadcastRoom(oldRoom, fmt.Sprintf("EVT ROOM PRESENCE LEAVE %s", player.Username), player.Username)
	manager.BroadcastRoom(roomID, fmt.Sprintf("EVT ROOM PRESENCE ENTER %s", player.Username), player.Username)
}

func (manager *WorldManager) NPCsInRoom(roomID string) []*model.NPC {
	return manager.npcs.ListInRoom(roomID)
}

func (manager *WorldManager) PlayersInRoom(roomID string) []string {
	return manager.presence.GetPlayersInRoom(roomID)
}

func (manager *WorldManager) TotalPlayers() int {
	return manager.presence.TotalPlayers()
}

func (manager *WorldManager) BroadcastRoom(roomID, message, excludedPlayer string) {
	manager.presence.BroadcastRoom(roomID, message, excludedPlayer)
}

func (manager *WorldManager) TakeItem(roomID, query string) (string, bool) {
	return manager.world.TakeItemFromRoom(roomID, query)
}

func (manager *WorldManager) DropItem(roomID, query string, inventory []string) (string, []string, bool) {
	return manager.world.DropItemToRoom(roomID, query, inventory)
}