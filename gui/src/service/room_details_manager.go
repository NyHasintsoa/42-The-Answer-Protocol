package service

import "fmt"

type RoomItemDetail struct {
	ID          string
	Name        string
	Description string
}

type RoomNPCDetail struct {
	Key         string
	Name        string
	Description string
	Kind        string
	HasQuest    bool
	QuestID     string
}

type RoomDetailsManager struct {
	IsOpen     bool
	RoomName   string
	Items      []RoomItemDetail
	NPCs       []RoomNPCDetail
}

func NewRoomDetailsManager() *RoomDetailsManager {
	return &RoomDetailsManager{
		IsOpen: true,
		Items:  []RoomItemDetail{},
		NPCs:   []RoomNPCDetail{},
	}
}

func (rm *RoomDetailsManager) SetRoomDetails(roomName string, items []RoomItemDetail, npcs []RoomNPCDetail) {
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

func (rm *RoomDetailsManager) TalkToNPC(npcKey string, npcName string) {
	fmt.Printf("[ACTION] Talking to NPC: %s (%s) in room '%s'\n", npcName, npcKey, rm.RoomName)
}

func (rm *RoomDetailsManager) OpenQuest(npcKey string, questID string) {
	fmt.Printf("[ACTION] Opened Quest '%s' from NPC: %s\n", questID, npcKey)
}

func (rm *RoomDetailsManager) Toggle() {
	rm.IsOpen = !rm.IsOpen
}