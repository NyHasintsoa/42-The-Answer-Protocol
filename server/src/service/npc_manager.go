package service

import (
	"math/rand"
	"strings"

	"tap-server/src/model"
)

type NpcManager struct {
	world *WorldService
}

type NPCRole string

const (
	NPCMonster NPCRole = "monster"
	NPCSeller  NPCRole = "seller"
	NPCPerson  NPCRole = "person"
)

func NewNpcManager(world *WorldService) *NpcManager {
	return &NpcManager{world: world}
}

func (manager *NpcManager) Get(npcID string) (*model.NPC, bool) {
	npc, ok := manager.world.GetNPC(npcID)
	if !ok {
		return nil, false
	}
	return npc, true
}

func (manager *NpcManager) FindInRoom(roomID, query string) (*model.NPC, bool) {
	return manager.world.FindNPCInRoom(roomID, query)
}

func (manager *NpcManager) ListInRoom(roomID string) []*model.NPC {
	room, ok := manager.world.GetRoom(roomID)
	if !ok {
		return nil
	}
	npcs := make([]*model.NPC, 0, len(room.NPCs))
	for _, npcID := range room.NPCs {
		if npc, ok := manager.Get(npcID); ok {
			npcs = append(npcs, npc)
		}
	}
	return npcs
}

func (manager *NpcManager) Role(npc *model.NPC) NPCRole {
	if npc == nil {
		return NPCPerson
	}
	if npc.CanBeAttacked {
		return NPCMonster
	}
	if strings.EqualFold(npc.ImageType, "merchant") || strings.EqualFold(npc.Type, "seller") {
		return NPCSeller
	}
	return NPCPerson
}

func (manager *NpcManager) IsMonster(npc *model.NPC) bool {
	return manager.Role(npc) == NPCMonster
}

func (manager *NpcManager) IsAlive(npcID string) bool {
	return manager.world.IsNPCAlive(npcID)
}

func (manager *NpcManager) Damage(npcID string, damage int) (int, bool) {
	return manager.world.DamageNPC(npcID, damage)
}

func (manager *NpcManager) Dialogue(npc *model.NPC) string {
	if npc == nil || len(npc.Dialogue) == 0 {
		return "Hello."
	}
	return npc.Dialogue[rand.Intn(len(npc.Dialogue))]
}