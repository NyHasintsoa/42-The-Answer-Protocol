package service

import (
	"strings"

	"tap-gui/src/model/npc"
)

type RoomNPCManager struct {
	npcsByRoom   map[string][]npc.NPCConfig
	activeRoomID string
}

// NewRoomNPCManager creates an NPC manager with no room characters.
func NewRoomNPCManager() *RoomNPCManager {
	return &RoomNPCManager{npcsByRoom: make(map[string][]npc.NPCConfig)}
}

// SetRoomNPCs replaces the NPC configurations stored for roomID with a copy of npcs.
// roomID identifies the room; npcs is the room's complete NPC list.
func (manager *RoomNPCManager) SetRoomNPCs(roomID string, npcs []npc.NPCConfig) {
	manager.npcsByRoom[roomID] = append([]npc.NPCConfig(nil), npcs...)
}

// SetActiveRoom selects roomID as the room used by active-NPC operations.
func (manager *RoomNPCManager) SetActiveRoom(roomID string) {
	manager.activeRoomID = roomID
}

// ActiveNPCs returns a copy of the selected room's NPC configurations.
func (manager *RoomNPCManager) ActiveNPCs() []npc.NPCConfig {
	return append([]npc.NPCConfig(nil), manager.npcsByRoom[manager.activeRoomID]...)
}

// UpdateActiveNPC replaces the active room's matching NPC configuration with updated.
// updated is matched by its key without regard to letter case; the method reports whether it found that NPC.
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

// DamageActiveNPC applies positive damage to the attackable active-room NPC identified by key.
// key is matched without regard to letter case; damage is the number of hit points to subtract.
// The returned configuration reflects the updated NPC; the boolean reports whether damage was applied.
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
