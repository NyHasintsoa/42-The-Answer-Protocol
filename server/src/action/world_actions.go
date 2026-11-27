package action

import (
	"encoding/json"
	"fmt"
	"strings"

	"tap-server/src/model"
)

func (r *ActionRouter) handleLook(player *model.Player) string {
	response, ok := r.WorldManager.Look(player)
	if !ok {
		return "ERR 500 ROOM_NOT_FOUND"
	}
	return jsonResponse(response)
}

func (r *ActionRouter) handleMove(player *model.Player, argument string) string {
	if argument == "" || strings.Contains(argument, " ") {
		return "ERR 400 INVALID_ARGUMENTS"
	}
	oldRoom := player.CurrentRoom
	if _, ok := r.WorldManager.Move(player, argument); !ok {
		return "ERR 301 NO_EXIT"
	}
	r.Store.SavePlayerSnapshot(player)
	r.Store.RecordInteraction("INFO", model.InteractionMove, player.Username, "ROOM", player.CurrentRoom, fmt.Sprintf("%s -> %s", oldRoom, player.CurrentRoom))
	return "OK room=" + player.CurrentRoom
}

func (r *ActionRouter) handleTake(player *model.Player, itemQuery string) string {
	if itemQuery == "" {
		return "ERR 400 INVALID_ARGUMENTS"
	}
	itemID, ok := r.InventoryManager.Take(player, itemQuery)
	if !ok {
		return "ERR 404 ITEM_NOT_FOUND"
	}
	r.Store.SavePlayerSnapshot(player)
	r.Store.RecordInteraction("INFO", model.InteractionItem, player.Username, "TAKE", player.CurrentRoom, itemID)
	r.WorldManager.BroadcastRoom(player.CurrentRoom, fmt.Sprintf("EVT ROOM ITEM TAKE %s %s", player.Username, itemID), player.Username)
	return "OK taken=" + itemID
}

func (r *ActionRouter) handleDrop(player *model.Player, itemQuery string) string {
	if itemQuery == "" {
		return "ERR 400 INVALID_ARGUMENTS"
	}
	itemID, ok := r.InventoryManager.Drop(player, itemQuery)
	if !ok {
		return "ERR 404 ITEM_NOT_IN_INVENTORY"
	}
	r.Store.SavePlayerSnapshot(player)
	r.Store.RecordInteraction("INFO", model.InteractionItem, player.Username, "DROP", player.CurrentRoom, itemID)
	r.WorldManager.BroadcastRoom(player.CurrentRoom, fmt.Sprintf("EVT ROOM ITEM DROP %s %s", player.Username, itemID), player.Username)
	return "OK dropped=" + itemID
}

func (r *ActionRouter) handleUse(player *model.Player, itemQuery string) string {
	if itemQuery == "" {
		return "ERR 400 INVALID_ARGUMENTS"
	}
	item, exists := r.WorldManager.Item(itemQuery)
	if !exists {
		return "ERR 404 ITEM_NOT_FOUND"
	}
	switch strings.ToLower(item.Category) {
	case "heal":
		if player.HP >= player.MaxHP {
			return "ERR 409 HEALTH_FULL"
		}
	case "attack":
	default:
		return "ERR 405 ITEM_NOT_USABLE"
	}
	if item.Cost <= 0 {
		return "ERR 422 INVALID_ITEM_EFFECT"
	}
	if !r.InventoryManager.Consume(player, item.ID) {
		return "ERR 404 ITEM_NOT_IN_INVENTORY"
	}

	effectAmount := item.Cost
	switch strings.ToLower(item.Category) {
	case "heal":
		if remaining := player.MaxHP - player.HP; effectAmount > remaining {
			effectAmount = remaining
		}
		player.HP += effectAmount
	case "attack":
		player.AttackBonus += effectAmount
	}

	result := map[string]any{
		"item_id":      item.ID,
		"name":         item.Name,
		"category":     strings.ToLower(item.Category),
		"amount":       effectAmount,
		"hp":           player.HP,
		"max_hp":       player.MaxHP,
		"attack_bonus": player.AttackBonus,
	}
	response, err := json.Marshal(result)
	if err != nil {
		return "ERR 500 INTERNAL_ERROR"
	}
	r.Store.SavePlayerSnapshot(player)
	r.Store.RecordInteraction("INFO", model.InteractionItem, player.Username, "USE", player.CurrentRoom, string(response))
	return "OK " + string(response)
}
