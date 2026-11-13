package action

import (
	"fmt"
	"unicode"

	"tap-server/src/model"
	"tap-server/src/service"
)

func (r *ActionRouter) Disconnect(sender service.ClientSender) {
	r.mu.Lock()
	defer r.mu.Unlock()
	player := sender.GetPlayer()
	if player == nil || player.Username == "" || !player.IsConnected {
		return
	}
	username := player.Username
	player.IsConnected = false
	r.Broadcast.UnregisterClient(username)
	r.Store.SavePlayerSnapshot(player)
	r.Store.RecordInteraction("INFO", model.InteractionSystem, username, "GLOBAL", player.CurrentRoom, "Disconnected")
}

func (r *ActionRouter) handleConnect(sender service.ClientSender, username string) string {
	player := sender.GetPlayer()
	if player == nil || player.IsConnected {
		return "ERR 400 ALREADY_CONNECTED"
	}
	if !validUsername(username) {
		return "ERR 400 INVALID_USERNAME"
	}
	if saved, ok := r.Store.GetSavedPlayer(username); ok {
		*player = *saved
		player.IsConnected = false
		if _, ok := r.WorldManager.Room(player.CurrentRoom); !ok {
			player.CurrentRoom = "start"
		}
		if player.HP <= 0 {
			player.HP = 50
		}
		if player.MaxHP <= 0 {
			player.MaxHP = 100
		}
		if player.Inventory == nil {
			player.Inventory = []string{}
		}
		if player.Quests == nil {
			player.Quests = make(map[string]*model.PlayerQuest)
		}
		player.InCombatWith = ""
		player.GroupID = ""
	} else {
		*player = model.Player{
			Username:    username,
			CurrentRoom: "start",
			HP:          100,
			MaxHP:       100,
			Inventory:   []string{},
			Quests:      make(map[string]*model.PlayerQuest),
		}
	}
	if !r.Broadcast.RegisterClient(username, sender) {
		return "ERR 201 NAME_IN_USE"
	}
	player.IsConnected = true
	r.Store.SavePlayerSnapshot(player)
	r.Store.RecordInteraction("INFO", model.InteractionSystem, username, "GLOBAL", player.CurrentRoom, "Connected")
	r.WorldManager.BroadcastRoom(player.CurrentRoom, fmt.Sprintf("EVT ROOM PRESENCE ENTER %s", username), username)
	return "OK connected"
}

func validUsername(username string) bool {
	if username == "" || len(username) > 32 {
		return false
	}
	for _, char := range username {
		if unicode.IsSpace(char) || unicode.IsControl(char) {
			return false
		}
	}
	return true
}
