package service

import (
	"strings"

	"tap-gui/src/model/npc"
)

type RoomNPCManager struct {
	npcsByRoom   map[string][]npc.NPCConfig
	activeRoomID string
}

func NewRoomNPCManager() *RoomNPCManager {
	return &RoomNPCManager{npcsByRoom: make(map[string][]npc.NPCConfig)}
}

func (manager *RoomNPCManager) SetRoomNPCs(roomID string, npcs []npc.NPCConfig) {
	manager.npcsByRoom[roomID] = append([]npc.NPCConfig(nil), npcs...)
}

func (manager *RoomNPCManager) SetActiveRoom(roomID string) {
	manager.activeRoomID = roomID
}

func (manager *RoomNPCManager) ActiveNPCs() []npc.NPCConfig {
	return append([]npc.NPCConfig(nil), manager.npcsByRoom[manager.activeRoomID]...)
}

func (manager *RoomNPCManager) UpdateActiveNPC(updated npc.NPCConfig) bool {
	npcs := manager.npcsByRoom[manager.activeRoomID]
	for index, existing := range npcs {
		if strings.EqualFold(existing.Key, updated.Key) {
			npcs[index] = updated
			manager.npcsByRoom[manager.activeRoomID] = npcs
			return true
		}
	}
	return false
}

func (manager *RoomNPCManager) DamageActiveNPC(key string, damage int) (npc.NPCConfig, bool) {
	for _, current := range manager.npcsByRoom[manager.activeRoomID] {
		if !strings.EqualFold(current.Key, key) || !current.CanBeAttacked || damage <= 0 || (current.MaxHP > 0 && current.HP <= 0) {
			continue
		}
		if current.MaxHP > 0 {
			current.HP -= damage
			if current.HP < 0 {
				current.HP = 0
			}
			if current.HP == 0 {
				current.CanBeAttacked = false
			}
		}
		manager.UpdateActiveNPC(current)
		return current, true
	}
	return npc.NPCConfig{}, false
}
