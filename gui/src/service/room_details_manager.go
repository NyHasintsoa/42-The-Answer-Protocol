package service

import (
	"fmt"
	"tap-gui/src/model/npc"
)

type RoomItemDetail struct {
	ID          string
	Name        string
	Description string
}

type RoomDetailsManager struct {
	IsOpen        bool
	RoomName      string
	Items         []RoomItemDetail
	NPCs          []npc.NPCConfig
	OnOpenActions func(npc npc.NPCConfig)
	OnAttackNPC   func(npc npc.NPCConfig)
}

func NewRoomDetailsManager() *RoomDetailsManager {
	return &RoomDetailsManager{
		IsOpen: true,
		Items:  []RoomItemDetail{},
		NPCs:   []npc.NPCConfig{},
	}
}

func (rm *RoomDetailsManager) SetRoomDetails(roomName string, items []RoomItemDetail, npcs []npc.NPCConfig) {
	rm.RoomName = roomName
	rm.Items = items
	rm.NPCs = npcs
}

func (rm *RoomDetailsManager) TakeItem(itemID string) {
	fmt.Printf("[ACTION] Take item: %s from room '%s'\n", itemID, rm.RoomName)
	for i, item := range rm.Items {
		if item.ID == itemID {
			rm.Items = append(rm.Items[:i], rm.Items[i+1:]...)
			break
		}
	}
}

func (rm *RoomDetailsManager) OpenActions(npc npc.NPCConfig) {
	if rm.OnOpenActions != nil {
		rm.OnOpenActions(npc)
	}
}

func (rm *RoomDetailsManager) AttackNPC(npc npc.NPCConfig) {
	if rm.OnAttackNPC != nil {
		rm.OnAttackNPC(npc)
	}
}

func (rm *RoomDetailsManager) Toggle() {
	rm.IsOpen = !rm.IsOpen
}