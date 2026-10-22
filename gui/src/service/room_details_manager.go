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

func NewRoomDetailsManager(itemManager *RoomItemManager, npcManager *RoomNPCManager) *RoomDetailsManager {
	return &RoomDetailsManager{
		IsOpen:      true,
		ItemManager: itemManager,
		NPCManager:  npcManager,
	}
}

func (manager *RoomDetailsManager) SetActiveRoom(roomID, roomName string) {
	manager.RoomName = roomName
	if manager.ItemManager != nil {
		manager.ItemManager.SetActiveRoom(roomID)
	}
	if manager.NPCManager != nil {
		manager.NPCManager.SetActiveRoom(roomID)
	}
}

func (manager *RoomDetailsManager) Items() []RoomItemDetail {
	if manager.ItemManager == nil {
		return nil
	}
	return manager.ItemManager.ActiveItems()
}

func (manager *RoomDetailsManager) NPCs() []npc.NPCConfig {
	if manager.NPCManager == nil {
		return nil
	}
	return manager.NPCManager.ActiveNPCs()
}

func (manager *RoomDetailsManager) TakeItem(itemID string) bool {
	return manager.ItemManager != nil && manager.ItemManager.RemoveItemFromActiveRoom(itemID)
}

func (manager *RoomDetailsManager) OpenActions(npcConfig npc.NPCConfig) {
	if manager.OnOpenActions != nil {
		manager.OnOpenActions(npcConfig)
	}
}

func (manager *RoomDetailsManager) AttackNPC(npcConfig npc.NPCConfig) {
	if manager.OnAttackNPC != nil {
		manager.OnAttackNPC(npcConfig)
	}
}

func (manager *RoomDetailsManager) Toggle() {
	manager.IsOpen = !manager.IsOpen
}
