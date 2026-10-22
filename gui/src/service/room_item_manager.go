package service

type RoomItemDetail struct {
	ID          string
	Name        string
	Description string
}

type RoomItemManager struct {
	itemsByRoom  map[string][]RoomItemDetail
	activeRoomID string
}

// NewRoomItemManager creates an item manager with no room items.
func NewRoomItemManager() *RoomItemManager {
	return &RoomItemManager{itemsByRoom: make(map[string][]RoomItemDetail)}
}

// SetRoomItems replaces the stored items for roomID with a copy of items.
// roomID identifies the room; items is the room's complete item list.
func (manager *RoomItemManager) SetRoomItems(roomID string, items []RoomItemDetail) {
	manager.itemsByRoom[roomID] = append([]RoomItemDetail(nil), items...)
}

// SetActiveRoom selects roomID as the room used by active-room operations.
func (manager *RoomItemManager) SetActiveRoom(roomID string) {
	manager.activeRoomID = roomID
}

// ActiveItems returns a copy of the selected room's items.
func (manager *RoomItemManager) ActiveItems() []RoomItemDetail {
	return append([]RoomItemDetail(nil), manager.itemsByRoom[manager.activeRoomID]...)
}

// AddItemToActiveRoom appends item to the active room and reports whether it was added.
// item must have a non-empty ID, and an active room must be selected.
func (manager *RoomItemManager) AddItemToActiveRoom(item RoomItemDetail) bool {
	if manager.activeRoomID == "" || item.ID == "" {
		return false
	}
	manager.itemsByRoom[manager.activeRoomID] = append(manager.itemsByRoom[manager.activeRoomID], item)
	return true
}

// RemoveItemFromActiveRoom removes the item with itemID from the active room.
// It returns true if a matching item was found and removed.
func (manager *RoomItemManager) RemoveItemFromActiveRoom(itemID string) bool {
	items := manager.itemsByRoom[manager.activeRoomID]
	for index, item := range items {
		if item.ID == itemID {
			manager.itemsByRoom[manager.activeRoomID] = append(items[:index], items[index+1:]...)
			return true
		}
	}
	return false
}
