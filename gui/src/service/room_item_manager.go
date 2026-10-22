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

func NewRoomItemManager() *RoomItemManager {
	return &RoomItemManager{itemsByRoom: make(map[string][]RoomItemDetail)}
}

func (manager *RoomItemManager) SetRoomItems(roomID string, items []RoomItemDetail) {
	manager.itemsByRoom[roomID] = append([]RoomItemDetail(nil), items...)
}

func (manager *RoomItemManager) SetActiveRoom(roomID string) {
	manager.activeRoomID = roomID
}

func (manager *RoomItemManager) ActiveItems() []RoomItemDetail {
	return append([]RoomItemDetail(nil), manager.itemsByRoom[manager.activeRoomID]...)
}

func (manager *RoomItemManager) AddItemToActiveRoom(item RoomItemDetail) bool {
	if manager.activeRoomID == "" || item.ID == "" {
		return false
	}
	manager.itemsByRoom[manager.activeRoomID] = append(manager.itemsByRoom[manager.activeRoomID], item)
	return true
}

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
