package service

import "tap-gui/src/model/npc"

type RoomDetailsManager struct {
	IsOpen        bool
	RoomName      string
	ItemManager   *RoomItemManager
	NPCManager    *RoomNPCManager
	OnOpenActions func(npc npc.NPCConfig)
	OnAttackNPC   func(npc npc.NPCConfig)
}

// NewRoomDetailsManager creates a room details manager backed by itemManager and npcManager.
// itemManager stores room items; npcManager stores room characters.
func NewRoomDetailsManager(itemManager *RoomItemManager, npcManager *RoomNPCManager) *RoomDetailsManager {
	return &RoomDetailsManager{
		IsOpen:      true,
		ItemManager: itemManager,
		NPCManager:  npcManager,
	}
}

// SetActiveRoom selects the room whose details are exposed by manager.
// roomID identifies the room in the item and NPC managers; roomName is its display name.
func (manager *RoomDetailsManager) SetActiveRoom(roomID, roomName string) {
	manager.RoomName = roomName
	if manager.ItemManager != nil {
		manager.ItemManager.SetActiveRoom(roomID)
	}
	if manager.NPCManager != nil {
		manager.NPCManager.SetActiveRoom(roomID)
	}
}

// Items returns a copy of the active room's item details, or nil when no item manager is configured.
func (manager *RoomDetailsManager) Items() []RoomItemDetail {
	if manager.ItemManager == nil {
		return nil
	}
	return manager.ItemManager.ActiveItems()
}

// NPCs returns a copy of the active room's NPC configurations, or nil when no NPC manager is configured.
func (manager *RoomDetailsManager) NPCs() []npc.NPCConfig {
	if manager.NPCManager == nil {
		return nil
	}
	return manager.NPCManager.ActiveNPCs()
}

// TakeItem removes the active room item identified by itemID and reports whether it was found.
func (manager *RoomDetailsManager) TakeItem(itemID string) bool {
	return manager.ItemManager != nil && manager.ItemManager.RemoveItemFromActiveRoom(itemID)
}

// OpenActions invokes the configured callback to open actions for npcConfig.
func (manager *RoomDetailsManager) OpenActions(npcConfig npc.NPCConfig) {
	if manager.OnOpenActions != nil {
		manager.OnOpenActions(npcConfig)
	}
}

// AttackNPC invokes the configured callback to start combat with npcConfig.
func (manager *RoomDetailsManager) AttackNPC(npcConfig npc.NPCConfig) {
	if manager.OnAttackNPC != nil {
		manager.OnAttackNPC(npcConfig)
	}
}

// Toggle changes whether the room details panel is open.
func (manager *RoomDetailsManager) Toggle() {
	manager.IsOpen = !manager.IsOpen
}
