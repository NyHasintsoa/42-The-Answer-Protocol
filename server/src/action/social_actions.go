package action

import (
	"fmt"
	"strings"

	"tap-server/src/model"
)

func (r *ActionRouter) handleChat(player *model.Player, argument string) string {
	parts := strings.SplitN(argument, " ", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[1]) == "" {
		return "ERR 400 INVALID_CHAT"
	}
	scope := strings.ToUpper(parts[0])
	message := strings.TrimSpace(parts[1])
	switch r.ChatManager.Send(player, scope, message) {
	case "NOT_IN_GROUP":
		return "ERR 401 NOT_IN_GROUP"
	case "INVALID_SCOPE":
		return "ERR 400 INVALID_SCOPE"
	}
	return "OK"
}

func (r *ActionRouter) handleGroup(player *model.Player, argument string) string {
	parts := strings.SplitN(argument, " ", 2)
	if len(parts) == 0 || parts[0] == "" {
		return "ERR 400 INVALID_GROUP_CMD"
	}
	action := strings.ToUpper(parts[0])
	target := ""
	if len(parts) == 2 {
		target = strings.TrimSpace(parts[1])
	}
	switch action {
	case "CREATE":
		if target != "" {
			return "ERR 400 INVALID_ARGUMENTS"
		}
		if player.GroupID != "" {
			return "ERR 402 ALREADY_IN_GROUP"
		}
		groupID, ok := r.Broadcast.CreateGroup(player.Username)
		if !ok {
			return "ERR 402 ALREADY_IN_GROUP"
		}
		return "OK group=" + groupID
	case "INVITE":
		if !singleArgument(target) {
			return "ERR 400 INVALID_ARGUMENTS"
		}
		if player.GroupID == "" {
			return "ERR 401 NOT_IN_GROUP"
		}
		if _, ok := r.Broadcast.InviteToGroup(player.Username, target); !ok {
			return "ERR 404 GROUP_NOT_FOUND"
		}
		r.Broadcast.SendToUser(target, fmt.Sprintf("EVT GROUP INVITE %s", player.Username))
		return "OK"
	case "JOIN":
		if !singleArgument(target) {
			return "ERR 400 INVALID_ARGUMENTS"
		}
		if player.GroupID != "" {
			return "ERR 402 ALREADY_IN_GROUP"
		}
		groupID, ok := r.Broadcast.JoinGroup(player.Username, target)
		if !ok {
			return "ERR 404 GROUP_NOT_FOUND"
		}
		r.Broadcast.BroadcastGroup(groupID, fmt.Sprintf("EVT GROUP JOIN %s", player.Username))
		return "OK group=" + groupID
	case "LEAVE":
		if target != "" {
			return "ERR 400 INVALID_ARGUMENTS"
		}
		if player.GroupID == "" {
			return "ERR 401 NOT_IN_GROUP"
		}
		groupID, _ := r.Broadcast.LeaveGroup(player.Username)
		r.Broadcast.BroadcastGroup(groupID, fmt.Sprintf("EVT GROUP LEAVE %s", player.Username))
		return "OK"
	default:
		return "ERR 400 INVALID_GROUP_CMD"
	}
}

func singleArgument(argument string) bool {
	return argument != "" && !strings.ContainsAny(argument, " \t")
}
