package service

import (
	"tap-server/src/model"
)

type QuestManager struct {
	world     *WorldService
	npcs      *NpcManager
	inventory *InventoryManager
}

type QuestProgress struct {
	Quest  *model.Quest
	Status string
}

func NewQuestManager(world *WorldService, npcs *NpcManager, inventory *InventoryManager) *QuestManager {
	return &QuestManager{world: world, npcs: npcs, inventory: inventory}
}

func (manager *QuestManager) Interact(player *model.Player, npcQuery string) (QuestProgress, string) {
	npc, ok := manager.npcs.FindInRoom(player.CurrentRoom, npcQuery)
	if !ok {
		return QuestProgress{}, "NPC_NOT_FOUND"
	}
	if npc.QuestID == "" {
		return QuestProgress{}, "NO_QUEST_AVAILABLE"
	}
	quest, ok := manager.world.GetQuest(npc.QuestID)
	if !ok {
		return QuestProgress{}, "NO_QUEST_AVAILABLE"
	}
	if player.Quests == nil {
		player.Quests = make(map[string]*model.PlayerQuest)
	}
	playerQuest, accepted := player.Quests[quest.ID]
	if accepted && playerQuest != nil && playerQuest.Status == model.QuestCompleted {
		return QuestProgress{}, "NO_QUEST_AVAILABLE"
	}

	status := "active"
	if !accepted || playerQuest == nil {
		playerQuest = &model.PlayerQuest{
			QuestID:  quest.ID,
			Status:   model.QuestActive,
			Progress: "0/1",
		}
		player.Quests[quest.ID] = playerQuest
		status = "available"
	}
	if manager.inventory.Consume(player, quest.RequiredItem) {
		manager.inventory.Add(player, quest.Reward)
		playerQuest.Status = model.QuestCompleted
		playerQuest.Progress = "1/1"
		status = "completed"
	}
	return QuestProgress{Quest: quest, Status: status}, ""
}

func (manager *QuestManager) List(player *model.Player) []*model.PlayerQuest {
	quests := make([]*model.PlayerQuest, 0, len(player.Quests))
	for _, quest := range player.Quests {
		if quest != nil {
			copyQuest := *quest
			quests = append(quests, &copyQuest)
		}
	}
	return quests
}

func (manager *QuestManager) Get(questID string) (*model.Quest, bool) {
	return manager.world.GetQuest(questID)
}