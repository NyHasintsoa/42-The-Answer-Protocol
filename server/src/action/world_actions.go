package action

import (
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
	return "OK dropped=" + itemID
}
