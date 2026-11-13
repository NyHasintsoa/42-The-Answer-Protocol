package service

import (
	"fmt"
	"strings"

	"tap-server/src/model"
	"tap-server/src/util"
)

type ChatManager struct {
	broadcast *BroadcastService
	store     *util.JSONStore
}

func NewChatManager(broadcast *BroadcastService, store *util.JSONStore) *ChatManager {
	return &ChatManager{broadcast: broadcast, store: store}
}

func (manager *ChatManager) Send(player *model.Player, scope, message string) string {
	switch strings.ToUpper(strings.TrimSpace(scope)) {
	case "GLOBAL":
		manager.store.RecordInteraction("INFO", model.InteractionChat, player.Username, "GLOBAL", "", message)
		manager.broadcast.BroadcastGlobal(fmt.Sprintf("EVT GLOBAL CHAT %s %s", player.Username, message))
	case "ROOM":
		manager.store.RecordInteraction("INFO", model.InteractionChat, player.Username, "ROOM", player.CurrentRoom, message)
		manager.broadcast.BroadcastRoom(player.CurrentRoom, fmt.Sprintf("EVT ROOM CHAT %s %s", player.Username, message), "")
	case "GROUP":
		if player.GroupID == "" {
			return "NOT_IN_GROUP"
		}
		manager.store.RecordInteraction("INFO", model.InteractionChat, player.Username, "GROUP", player.GroupID, message)
		manager.broadcast.BroadcastGroup(player.GroupID, fmt.Sprintf("EVT GROUP CHAT %s %s", player.Username, message))
	default:
		return "INVALID_SCOPE"
	}
	return ""
}