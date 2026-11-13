package action

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"strings"

	"tap-server/src/model"
)

func (r *ActionRouter) handleTalk(player *model.Player, npcQuery string) string {
	if npcQuery == "" {
		return "ERR 400 INVALID_ARGUMENTS"
	}
	npc, ok := r.NpcManager.FindInRoom(player.CurrentRoom, npcQuery)
	if !ok {
		return "ERR 404 NPC_NOT_FOUND"
	}
	dialogue := r.NpcManager.Dialogue(npc)
	r.Store.RecordInteraction("INFO", model.InteractionSystem, player.Username, "TALK", npc.ID, dialogue)
	return "OK " + dialogue
}

func (r *ActionRouter) handleAttack(player *model.Player, npcQuery string) string {
	if npcQuery == "" {
		return "ERR 400 INVALID_ARGUMENTS"
	}
	npc, ok := r.NpcManager.FindInRoom(player.CurrentRoom, npcQuery)
	if !ok {
		return "ERR 404 NPC_NOT_FOUND"
	}
	if !r.NpcManager.IsMonster(npc) || !r.NpcManager.IsAlive(npc.ID) {
		return "ERR 405 NPC_NOT_HOSTILE"
	}
	battleRoom := player.CurrentRoom
	damage := 15 + player.AttackBonus + rand.Intn(10)
	targetHP, ok := r.NpcManager.Damage(npc.ID, damage)
	if !ok {
		return "ERR 405 NPC_NOT_HOSTILE"
	}
	status := "combat"
	if targetHP == 0 {
		status = "victory"
		player.InCombatWith = ""
	} else {
		player.InCombatWith = npc.ID
		player.HP -= npc.AttackDamage
		if player.HP <= 0 {
			player.HP = 50
			status = "respawned"
		}
	}
	if status == "respawned" {
		r.WorldManager.Relocate(player, "start")
	}
	result := map[string]any{
		"attacker_hp": player.HP,
		"target_hp":   targetHP,
		"damage":      damage,
		"status":      status,
	}
	response, err := json.Marshal(result)
	if err != nil {
		return "ERR 500 INTERNAL_ERROR"
	}
	r.Store.SavePlayerSnapshot(player)
	r.Store.RecordInteraction("INFO", model.InteractionCombat, player.Username, "ATTACK", npc.ID, string(response))
	r.WorldManager.BroadcastRoom(battleRoom, fmt.Sprintf("EVT ROOM COMBAT %s attacked %s damage=%d target_hp=%d status=%s", player.Username, npc.Name, damage, targetHP, status), player.Username)
	return "OK " + string(response)
}

func (r *ActionRouter) handleDefend(player *model.Player) string {
	if player.InCombatWith == "" {
		return "ERR 405 NOT_IN_COMBAT"
	}
	npc, ok := r.NpcManager.FindInRoom(player.CurrentRoom, player.InCombatWith)
	if !ok || !r.NpcManager.IsAlive(npc.ID) {
		player.InCombatWith = ""
		return "ERR 404 NPC_NOT_FOUND"
	}
	damage := npc.AttackDamage / 2
	if damage < 1 {
		damage = 1
	}
	battleRoom := player.CurrentRoom
	player.HP -= damage
	status := "defending"
	if player.HP <= 0 {
		player.HP = 50
		status = "respawned"
	} else {
		player.InCombatWith = npc.ID
	}
	if status == "respawned" {
		r.WorldManager.Relocate(player, "start")
	}
	result := map[string]any{"hp": player.HP, "damage": damage, "status": status}
	response, err := json.Marshal(result)
	if err != nil {
		return "ERR 500 INTERNAL_ERROR"
	}
	r.Store.SavePlayerSnapshot(player)
	r.Store.RecordInteraction("INFO", model.InteractionCombat, player.Username, "DEFEND", npc.ID, string(response))
	r.WorldManager.BroadcastRoom(battleRoom, fmt.Sprintf("EVT ROOM COMBAT %s defended against %s damage=%d status=%s", player.Username, npc.Name, damage, status), player.Username)
	return "OK " + string(response)
}

func (r *ActionRouter) handleFlee(player *model.Player, direction string) string {
	if player.InCombatWith == "" {
		return "ERR 405 NOT_IN_COMBAT"
	}
	if direction == "" || strings.Contains(direction, " ") {
		return "ERR 400 INVALID_ARGUMENTS"
	}
	battleRoom := player.CurrentRoom
	room, ok := r.WorldManager.Room(battleRoom)
	if !ok {
		return "ERR 301 NO_EXIT"
	}
	destination, ok := room.Exits[strings.ToLower(direction)]
	if !ok {
		return "ERR 301 NO_EXIT"
	}
	npc, ok := r.NpcManager.FindInRoom(battleRoom, player.InCombatWith)
	if !ok || !r.NpcManager.IsAlive(npc.ID) {
		player.InCombatWith = ""
		return "ERR 404 NPC_NOT_FOUND"
	}
	damage := npc.AttackDamage
	player.HP -= damage
	status := "fled"
	if player.HP <= 0 {
		player.HP = 50
		status = "respawned"
	} else {
		r.WorldManager.Relocate(player, destination)
	}
	if status == "respawned" {
		r.WorldManager.Relocate(player, "start")
	}
	result := map[string]any{"hp": player.HP, "damage": damage, "status": status}
	response, err := json.Marshal(result)
	if err != nil {
		return "ERR 500 INTERNAL_ERROR"
	}
	r.Store.SavePlayerSnapshot(player)
	r.Store.RecordInteraction("INFO", model.InteractionCombat, player.Username, "FLEE", npc.ID, string(response))
	r.WorldManager.BroadcastRoom(battleRoom, fmt.Sprintf("EVT ROOM COMBAT %s fled from %s damage=%d status=%s", player.Username, npc.Name, damage, status), player.Username)
	return "OK " + string(response)
}

func (r *ActionRouter) handleStatus(player *model.Player) string {
	status := "healthy"
	if player.InCombatWith != "" {
		status = "combat"
	}
	return jsonResponse(map[string]any{
		"hp":     player.HP,
		"max_hp": player.MaxHP,
		"status": status,
	})
}

func (r *ActionRouter) handleQuest(player *model.Player, npcQuery string) string {
	if npcQuery == "" {
		return "ERR 400 INVALID_ARGUMENTS"
	}
	progress, errorCode := r.QuestManager.Interact(player, npcQuery)
	if errorCode != "" {
		switch errorCode {
		case "NPC_NOT_FOUND":
			return "ERR 404 NPC_NOT_FOUND"
		case "NO_QUEST_AVAILABLE":
			return "ERR 406 NO_QUEST_AVAILABLE"
		}
		return "ERR 500 INTERNAL_ERROR"
	}
	if progress.Status == "available" {
		r.Store.RecordInteraction("INFO", model.InteractionQuest, player.Username, "ACCEPT", progress.Quest.ID, progress.Quest.Title)
	}
	if progress.Status == "completed" {
		r.Store.RecordInteraction("INFO", model.InteractionQuest, player.Username, "COMPLETE", progress.Quest.ID, "reward="+progress.Quest.Reward)
	}
	r.Store.SavePlayerSnapshot(player)
	response := map[string]string{
		"quest_id":    progress.Quest.ID,
		"description": progress.Quest.Description,
		"reward":      progress.Quest.Reward,
		"status":      progress.Status,
	}
	return jsonResponse(response)
}

func (r *ActionRouter) handleQuests(player *model.Player) string {
	quests := r.QuestManager.List(player)
	sort.Slice(quests, func(i, j int) bool { return quests[i].QuestID < quests[j].QuestID })
	return jsonResponse(quests)
}
