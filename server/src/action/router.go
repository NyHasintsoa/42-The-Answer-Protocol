package action

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"unicode"

	"tap-server/src/model"
	"tap-server/src/service"
	"tap-server/src/util"
)

type ActionRouter struct {
	mu               sync.Mutex
	Broadcast        *service.BroadcastService
	Store            *util.JSONStore
	WorldManager     *service.WorldManager
	InventoryManager *service.InventoryManager
	ChatManager      *service.ChatManager
	QuestManager     *service.QuestManager
	NpcManager       *service.NpcManager
}

func NewActionRouter(world *service.WorldService, broadcast *service.BroadcastService, store *util.JSONStore) *ActionRouter {
	npcManager := service.NewNpcManager(world)
	worldManager := service.NewWorldManager(world, broadcast, npcManager)
	inventoryManager := service.NewInventoryManager(worldManager)
	return &ActionRouter{
		Broadcast:        broadcast,
		Store:            store,
		WorldManager:     worldManager,
		InventoryManager: inventoryManager,
		ChatManager:      service.NewChatManager(broadcast, store),
		QuestManager:     service.NewQuestManager(world, npcManager, inventoryManager),
		NpcManager:       npcManager,
	}
}

func (r *ActionRouter) Dispatch(sender service.ClientSender, rawLine string) (string, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	line := strings.TrimSpace(rawLine)
	if line == "" {
		return "ERR 400 EMPTY_COMMAND", false
	}
	for _, char := range line {
		if unicode.IsControl(char) {
			return "ERR 400 INVALID_COMMAND", false
		}
	}
	parts := strings.SplitN(line, " ", 2)
	command := strings.ToUpper(parts[0])
	argument := ""
	if len(parts) == 2 {
		argument = strings.TrimSpace(parts[1])
	}
	player := sender.GetPlayer()

	if command == "CONNECT" {
		return r.handleConnect(sender, argument), false
	}
	if command == "QUIT" {
		if player == nil || !player.IsConnected {
			return "ERR 401 NOT_AUTHENTICATED", false
		}
		return "OK bye", true
	}
	if player == nil || !player.IsConnected {
		return "ERR 401 NOT_AUTHENTICATED", false
	}

	switch command {
	case "LOOK":
		if argument != "" {
			return "ERR 400 INVALID_ARGUMENTS", false
		}
		return r.handleLook(player), false
	case "MOVE":
		return r.handleMove(player, argument), false
	case "CHAT":
		return r.handleChat(player, argument), false
	case "WHO":
		if argument != "" {
			return "ERR 400 INVALID_ARGUMENTS", false
		}
		return fmt.Sprintf("OK players=%d", r.WorldManager.TotalPlayers()), false
	case "GROUP":
		return r.handleGroup(player, argument), false
	case "TAKE":
		return r.handleTake(player, argument), false
	case "DROP":
		return r.handleDrop(player, argument), false
	case "INVENTORY":
		if argument != "" {
			return "ERR 400 INVALID_ARGUMENTS", false
		}
		return jsonResponse(r.InventoryManager.List(player)), false
	case "TALK":
		return r.handleTalk(player, argument), false
	case "ATTACK":
		return r.handleAttack(player, argument), false
	case "DEFEND":
		if argument != "" {
			return "ERR 400 INVALID_ARGUMENTS", false
		}
		return r.handleDefend(player), false
	case "FLEE":
		return r.handleFlee(player, argument), false
	case "STATUS":
		if argument != "" {
			return "ERR 400 INVALID_ARGUMENTS", false
		}
		return r.handleStatus(player), false
	case "QUEST":
		return r.handleQuest(player, argument), false
	case "QUESTS":
		if argument != "" {
			return "ERR 400 INVALID_ARGUMENTS", false
		}
		return r.handleQuests(player), false
	default:
		return "ERR 400 UNKNOWN_COMMAND", false
	}
}

func (r *ActionRouter) RecordSessionEvent(level, event, actor, remote, payload string) {
	r.Store.RecordEvent(model.InteractionRecord{
		Level:   level,
		Type:    model.InteractionSystem,
		Event:   event,
		Actor:   actor,
		Remote:  remote,
		Payload: payload,
	})
}

func (r *ActionRouter) RecordCommand(actor, remote, command, parameters string) {
	r.Store.RecordEvent(model.InteractionRecord{
		Level:   "INFO",
		Type:    model.InteractionSystem,
		Event:   "command_received",
		Actor:   actor,
		Remote:  remote,
		Command: command,
		Payload: parameters,
	})
}

func jsonResponse(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return "ERR 500 INTERNAL_ERROR"
	}
	return "OK " + string(data)
}
