package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"tap-server/src/model"
)

type WorldService struct {
	mu     sync.RWMutex
	config model.WorldConfig
}

func NewWorldService(mapPath string) (*WorldService, error) {
	raw, err := os.ReadFile(mapPath)
	if err != nil {
		return nil, fmt.Errorf("read world map %q: %w", mapPath, err)
	}

	var cfg model.WorldConfig
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("parse world map %q: %w", mapPath, err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("parse world map %q: multiple JSON values", mapPath)
		}
		return nil, fmt.Errorf("parse world map %q: %w", mapPath, err)
	}
	if err := validateWorld(&cfg); err != nil {
		return nil, fmt.Errorf("validate world map %q: %w", mapPath, err)
	}

	for id, room := range cfg.World.Rooms {
		room.ID = id
	}
	for id, npc := range cfg.NPCs {
		npc.ID = id
		if npc.HP == 0 && npc.MaxHP > 0 {
			npc.HP = npc.MaxHP
		}
	}
	for id, item := range cfg.Items {
		if item.ID == "" {
			item.ID = id
		}
	}
	for id, quest := range cfg.Quests {
		if quest.ID == "" {
			quest.ID = id
		}
	}

	return &WorldService{config: cfg}, nil
}

func validateWorld(cfg *model.WorldConfig) error {
	if cfg.World.Rooms == nil || cfg.NPCs == nil || cfg.Items == nil || cfg.Quests == nil {
		return fmt.Errorf("world.rooms, npcs, items, and quests are required")
	}
	if len(cfg.World.Rooms) < 8 {
		return fmt.Errorf("world must contain at least 8 rooms; found %d", len(cfg.World.Rooms))
	}
	if _, ok := cfg.World.Rooms["start"]; !ok {
		return fmt.Errorf("world.rooms must contain the start room")
	}

	obtainableItems := make(map[string]bool)
	for roomID, room := range cfg.World.Rooms {
		if room == nil {
			return fmt.Errorf("room %q is null", roomID)
		}
		if strings.TrimSpace(room.Name) == "" || strings.TrimSpace(room.Description) == "" {
			return fmt.Errorf("room %q must have a name and description", roomID)
		}
		for direction, destination := range room.Exits {
			if strings.TrimSpace(direction) == "" {
				return fmt.Errorf("room %q has an empty exit direction", roomID)
			}
			if _, ok := cfg.World.Rooms[destination]; !ok {
				return fmt.Errorf("invalid exit %s -> %s in room %s", direction, destination, roomID)
			}
		}
		for _, itemID := range room.Items {
			if _, ok := cfg.Items[itemID]; !ok {
				return fmt.Errorf("room %q references unknown item %q", roomID, itemID)
			}
			obtainableItems[itemID] = true
		}
		for _, npcID := range room.NPCs {
			if _, ok := cfg.NPCs[npcID]; !ok {
				return fmt.Errorf("room %q references unknown NPC %q", roomID, npcID)
			}
		}
	}
	if len(cfg.NPCs) < 3 || len(cfg.Items) < 4 || len(cfg.Quests) < 2 {
		return fmt.Errorf("world must contain at least 3 NPCs, 4 items, and 2 quests")
	}
	if len(obtainableItems) < 2 {
		return fmt.Errorf("world must place at least 2 distinct items in rooms")
	}

	visited := map[string]bool{"start": true}
	queue := []string{"start"}
	for len(queue) > 0 {
		roomID := queue[0]
		queue = queue[1:]
		for _, destination := range cfg.World.Rooms[roomID].Exits {
			if !visited[destination] {
				visited[destination] = true
				queue = append(queue, destination)
			}
		}
	}
	if len(visited) != len(cfg.World.Rooms) {
		return fmt.Errorf("all rooms must be reachable from start; reached %d of %d", len(visited), len(cfg.World.Rooms))
	}
	if !hasRoomLoop(cfg.World.Rooms) {
		return fmt.Errorf("world rooms must contain at least one loop")
	}
	hasBranch := false
	for roomID, room := range cfg.World.Rooms {
		if roomID != "start" && len(room.Exits) == 1 {
			hasBranch = true
			break
		}
	}
	if !hasBranch {
		return fmt.Errorf("world rooms must contain an optional branch")
	}

	for npcID, npc := range cfg.NPCs {
		if npc == nil || strings.TrimSpace(npc.Name) == "" {
			return fmt.Errorf("NPC %q must have a name", npcID)
		}
		if npc.CanBeAttacked && (npc.MaxHP <= 0 || npc.AttackDamage < 0) {
			return fmt.Errorf("hostile NPC %q must have positive max_hp and non-negative attack_damage", npcID)
		}
		if npc.HP < 0 || npc.HP > npc.MaxHP {
			return fmt.Errorf("NPC %q hp must be between 0 and max_hp", npcID)
		}
		if npc.QuestID != "" {
			quest, ok := cfg.Quests[npc.QuestID]
			if !ok || quest == nil || quest.Giver != npcID {
				return fmt.Errorf("NPC %q references invalid quest %q", npcID, npc.QuestID)
			}
		}
	}

	for itemID, item := range cfg.Items {
		if item == nil || strings.TrimSpace(item.Name) == "" {
			return fmt.Errorf("item %q must have a name", itemID)
		}
		if item.ID != "" && item.ID != itemID {
			return fmt.Errorf("item key %q does not match its id %q", itemID, item.ID)
		}
	}
	for questID, quest := range cfg.Quests {
		if quest == nil || strings.TrimSpace(quest.Title) == "" || quest.Giver == "" {
			return fmt.Errorf("quest %q must have a title and giver", questID)
		}
		if quest.ID != "" && quest.ID != questID {
			return fmt.Errorf("quest key %q does not match its id %q", questID, quest.ID)
		}
		giver, ok := cfg.NPCs[quest.Giver]
		if !ok || giver == nil || giver.QuestID != questID {
			return fmt.Errorf("quest %q has invalid giver %q", questID, quest.Giver)
		}
		if _, ok := cfg.Items[quest.RequiredItem]; !ok {
			return fmt.Errorf("quest %q references unknown required item %q", questID, quest.RequiredItem)
		}
		if _, ok := cfg.Items[quest.Reward]; !ok {
			return fmt.Errorf("quest %q references unknown reward item %q", questID, quest.Reward)
		}
	}
	return nil
}

func hasRoomLoop(rooms map[string]*model.Room) bool {
	state := make(map[string]uint8, len(rooms))
	var visit func(string) bool
	visit = func(roomID string) bool {
		state[roomID] = 1
		for _, destination := range rooms[roomID].Exits {
			if state[destination] == 1 || (state[destination] == 0 && visit(destination)) {
				return true
			}
		}
		state[roomID] = 2
		return false
	}
	for roomID := range rooms {
		if state[roomID] == 0 && visit(roomID) {
			return true
		}
	}
	return false
}

func (ws *WorldService) GetRoom(roomID string) (*model.Room, bool) {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	room, ok := ws.config.World.Rooms[roomID]
	if !ok {
		return nil, false
	}
	copyRoom := *room
	copyRoom.Exits = cloneStringMap(room.Exits)
	copyRoom.Items = append([]string(nil), room.Items...)
	copyRoom.NPCs = append([]string(nil), room.NPCs...)
	return &copyRoom, true
}

func (ws *WorldService) TakeItemFromRoom(roomID, query string) (string, bool) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	room, ok := ws.config.World.Rooms[roomID]
	if !ok {
		return "", false
	}
	match := strings.ToLower(strings.TrimSpace(query))
	for i, itemID := range room.Items {
		item := ws.config.Items[itemID]
		if strings.EqualFold(itemID, match) || (item != nil && strings.EqualFold(item.Name, match)) {
			room.Items = append(room.Items[:i], room.Items[i+1:]...)
			return itemID, true
		}
	}
	return "", false
}

func (ws *WorldService) DropItemToRoom(roomID, query string, inventory []string) (string, []string, bool) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	room, ok := ws.config.World.Rooms[roomID]
	if !ok {
		return "", inventory, false
	}
	match := strings.TrimSpace(query)
	for i, itemID := range inventory {
		item := ws.config.Items[itemID]
		if strings.EqualFold(itemID, match) || (item != nil && strings.EqualFold(item.Name, match)) {
			newInventory := append([]string(nil), inventory[:i]...)
			newInventory = append(newInventory, inventory[i+1:]...)
			room.Items = append(room.Items, itemID)
			return itemID, newInventory, true
		}
	}
	return "", inventory, false
}

func (ws *WorldService) FindNPCInRoom(roomID, query string) (*model.NPC, bool) {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	room, ok := ws.config.World.Rooms[roomID]
	if !ok {
		return nil, false
	}
	for _, npcID := range room.NPCs {
		npc := ws.config.NPCs[npcID]
		if strings.EqualFold(npcID, strings.TrimSpace(query)) || strings.EqualFold(npc.Name, strings.TrimSpace(query)) {
			copyNPC := *npc
			copyNPC.Dialogue = append([]string(nil), npc.Dialogue...)
			return &copyNPC, true
		}
	}
	return nil, false
}

func (ws *WorldService) GetNPC(npcID string) (*model.NPC, bool) {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	npc, ok := ws.config.NPCs[npcID]
	if !ok {
		return nil, false
	}
	copyNPC := *npc
	copyNPC.Dialogue = append([]string(nil), npc.Dialogue...)
	return &copyNPC, true
}

func (ws *WorldService) DamageNPC(npcID string, damage int) (int, bool) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	npc, ok := ws.config.NPCs[npcID]
	if !ok || npc.HP <= 0 {
		return 0, false
	}
	npc.HP -= damage
	if npc.HP < 0 {
		npc.HP = 0
	}
	return npc.HP, true
}

func (ws *WorldService) IsNPCAlive(npcID string) bool {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	npc, ok := ws.config.NPCs[npcID]
	return ok && npc.HP > 0
}

func (ws *WorldService) GetQuest(questID string) (*model.Quest, bool) {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	quest, ok := ws.config.Quests[questID]
	if !ok {
		return nil, false
	}
	copyQuest := *quest
	return &copyQuest, true
}

func (ws *WorldService) ItemName(itemID string) string {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	if item := ws.config.Items[itemID]; item != nil {
		return item.Name
	}
	return itemID
}

func cloneStringMap(src map[string]string) map[string]string {
	dst := make(map[string]string, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}
