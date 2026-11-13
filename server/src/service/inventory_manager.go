package service

import (
	"tap-server/src/model"
)

type InventoryManager struct {
	world *WorldManager
}

func NewInventoryManager(world *WorldManager) *InventoryManager {
	return &InventoryManager{world: world}
}

func (manager *InventoryManager) List(player *model.Player) []string {
	return append([]string(nil), player.Inventory...)
}

func (manager *InventoryManager) Take(player *model.Player, query string) (string, bool) {
	itemID, ok := manager.world.TakeItem(player.CurrentRoom, query)
	if !ok {
		return "", false
	}
	player.Inventory = append(player.Inventory, itemID)
	return itemID, true
}

func (manager *InventoryManager) Drop(player *model.Player, query string) (string, bool) {
	itemID, inventory, ok := manager.world.DropItem(player.CurrentRoom, query, player.Inventory)
	if !ok {
		return "", false
	}
	player.Inventory = inventory
	return itemID, true
}

func (manager *InventoryManager) Consume(player *model.Player, itemID string) bool {
	for index, heldItem := range player.Inventory {
		if heldItem != itemID {
			continue
		}
		player.Inventory = append(player.Inventory[:index:index], player.Inventory[index+1:]...)
		return true
	}
	return false
}

func (manager *InventoryManager) Add(player *model.Player, itemID string) {
	player.Inventory = append(player.Inventory, itemID)
}