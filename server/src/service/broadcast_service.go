package service

import (
	"fmt"
	"sort"
	"sync"

	"tap-server/src/model"
)

type ClientSender interface {
	SendLine(line string)
	GetPlayer() *model.Player
}

type BroadcastService struct {
	mu      sync.RWMutex
	clients map[string]ClientSender
	groups  map[string]*model.Group
}

func NewBroadcastService() *BroadcastService {
	return &BroadcastService{
		clients: make(map[string]ClientSender),
		groups:  make(map[string]*model.Group),
	}
}

func (bs *BroadcastService) RegisterClient(username string, sender ClientSender) bool {
	bs.mu.Lock()
	if _, exists := bs.clients[username]; exists {
		bs.mu.Unlock()
		return false
	}
	bs.clients[username] = sender
	count := len(bs.clients)
	bs.mu.Unlock()
	bs.BroadcastGlobal(fmt.Sprintf("EVT STATS players=%d", count))
	return true
}

func (bs *BroadcastService) UnregisterClient(username string) {
	bs.mu.Lock()
	sender, exists := bs.clients[username]
	if !exists {
		bs.mu.Unlock()
		return
	}
	player := sender.GetPlayer()
	delete(bs.clients, username)
	for id, group := range bs.groups {
		delete(group.Invites, username)
		if !group.Members[username] {
			continue
		}
		delete(group.Members, username)
		if group.Leader == username {
			members := sortedKeys(group.Members)
			if len(members) == 0 {
				delete(bs.groups, id)
			} else {
				group.Leader = members[0]
			}
		} else if len(group.Members) == 0 {
			delete(bs.groups, id)
		}
	}
	count := len(bs.clients)
	bs.mu.Unlock()

	if player != nil {
		bs.BroadcastRoom(player.CurrentRoom, fmt.Sprintf("EVT ROOM PRESENCE LEAVE %s", username), username)
	}
	bs.BroadcastGlobal(fmt.Sprintf("EVT STATS players=%d", count))
}

func (bs *BroadcastService) BroadcastGlobal(message string) {
	for _, client := range bs.clientSnapshot() {
		client.SendLine(message)
	}
}

func (bs *BroadcastService) BroadcastRoom(roomID, message, excludeUser string) {
	clients := bs.clientSnapshot()
	for username, client := range clients {
		if username == excludeUser {
			continue
		}
		player := client.GetPlayer()
		if player != nil && player.CurrentRoom == roomID {
			client.SendLine(message)
		}
	}
}

func (bs *BroadcastService) BroadcastGroup(groupID, message string) {
	bs.mu.RLock()
	group, ok := bs.groups[groupID]
	members := make([]string, 0)
	if ok {
		for member := range group.Members {
			members = append(members, member)
		}
	}
	sort.Strings(members)
	clients := make(map[string]ClientSender, len(members))
	for _, member := range members {
		if client, online := bs.clients[member]; online {
			clients[member] = client
		}
	}
	bs.mu.RUnlock()
	for _, member := range members {
		if client := clients[member]; client != nil {
			client.SendLine(message)
		}
	}
}

func (bs *BroadcastService) SendToUser(username, message string) bool {
	bs.mu.RLock()
	client, ok := bs.clients[username]
	bs.mu.RUnlock()
	if !ok {
		return false
	}
	client.SendLine(message)
	return true
}

func (bs *BroadcastService) GetPlayersInRoom(roomID string) []string {
	players := make([]string, 0)
	for username, client := range bs.clientSnapshot() {
		player := client.GetPlayer()
		if player != nil && player.CurrentRoom == roomID {
			players = append(players, username)
		}
	}
	sort.Strings(players)
	return players
}

func (bs *BroadcastService) TotalPlayers() int {
	bs.mu.RLock()
	defer bs.mu.RUnlock()
	return len(bs.clients)
}

func (bs *BroadcastService) CreateGroup(username string) (string, bool) {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	client, ok := bs.clients[username]
	if !ok || client.GetPlayer().GroupID != "" {
		return "", false
	}
	groupID := "grp_" + username
	if _, exists := bs.groups[groupID]; exists {
		return "", false
	}
	bs.groups[groupID] = &model.Group{
		ID:      groupID,
		Leader:  username,
		Members: map[string]bool{username: true},
		Invites: make(map[string]bool),
	}
	client.GetPlayer().GroupID = groupID
	return groupID, true
}

func (bs *BroadcastService) InviteToGroup(username, target string) (string, bool) {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	client, ok := bs.clients[username]
	if !ok || client.GetPlayer().GroupID == "" {
		return "", false
	}
	group := bs.groups[client.GetPlayer().GroupID]
	if group == nil || group.Leader != username || target == username {
		return "", false
	}
	if _, online := bs.clients[target]; !online {
		return "", false
	}
	targetPlayer := bs.clients[target].GetPlayer()
	if targetPlayer.GroupID != "" {
		return "", false
	}
	group.Invites[target] = true
	return group.ID, true
}

func (bs *BroadcastService) JoinGroup(username, leader string) (string, bool) {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	client, ok := bs.clients[username]
	if !ok || client.GetPlayer().GroupID != "" {
		return "", false
	}
	groupID := "grp_" + leader
	group := bs.groups[groupID]
	if group == nil || !group.Invites[username] {
		return "", false
	}
	delete(group.Invites, username)
	group.Members[username] = true
	client.GetPlayer().GroupID = groupID
	return groupID, true
}

func (bs *BroadcastService) LeaveGroup(username string) (string, bool) {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	client, ok := bs.clients[username]
	if !ok || client.GetPlayer().GroupID == "" {
		return "", false
	}
	groupID := client.GetPlayer().GroupID
	group := bs.groups[groupID]
	if group != nil {
		delete(group.Members, username)
		delete(group.Invites, username)
		if group.Leader == username {
			members := sortedKeys(group.Members)
			if len(members) > 0 {
				group.Leader = members[0]
			}
		}
		if len(group.Members) == 0 {
			delete(bs.groups, groupID)
		}
	}
	client.GetPlayer().GroupID = ""
	return groupID, true
}

func (bs *BroadcastService) clientSnapshot() map[string]ClientSender {
	bs.mu.RLock()
	defer bs.mu.RUnlock()
	clients := make(map[string]ClientSender, len(bs.clients))
	for username, client := range bs.clients {
		clients[username] = client
	}
	return clients
}

func sortedKeys(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
