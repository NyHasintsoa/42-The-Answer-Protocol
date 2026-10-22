package service

import (
	"fmt"
	"tap-gui/src/model/npc"
)

type Quest struct {
	ID              string
	Title           string
	NPCName         string
	CurrentProgress int
	TargetProgress  int
}

// IsCompleted reports whether the quest has reached its target progress.
func (q *Quest) IsCompleted() bool {
	return q.CurrentProgress >= q.TargetProgress
}

type QuestManager struct {
	IsOpen bool
	Quests []*Quest
}

// NewQuestManager creates a quest manager and initializes its default quests.
func NewQuestManager() *QuestManager {
	qm := &QuestManager{
		IsOpen: false,
		Quests: []*Quest{},
	}
	qm.initDefaultQuests()
	return qm
}

// initDefaultQuests replaces the current quest list with the built-in sample quests.
func (qm *QuestManager) initDefaultQuests() {
	qm.Quests = []*Quest{
		{ID: "npc_1", Title: "Talk to Guard Captain", NPCName: "Guard", CurrentProgress: 1, TargetProgress: 1},
		{ID: "npc_2", Title: "Gather 5 Herbs for Merchant", NPCName: "Merchant", CurrentProgress: 2, TargetProgress: 5},
		{ID: "npc_3", Title: "Clear 3 Slimes near Gate", NPCName: "Guard", CurrentProgress: 1, TargetProgress: 3},
		{ID: "npc_4", Title: "Deliver Package to Villager", NPCName: "Villager", CurrentProgress: 0, TargetProgress: 1},
		{ID: "npc_5", Title: "Investigate Strange Noises", NPCName: "Elder", CurrentProgress: 0, TargetProgress: 1},
		{ID: "npc_5", Title: "Investigate Strange Noises", NPCName: "Elder", CurrentProgress: 0, TargetProgress: 1},
		{ID: "npc_5", Title: "Investigate Strange Noises", NPCName: "Elder", CurrentProgress: 0, TargetProgress: 1},
		{ID: "npc_5", Title: "Investigate Strange Noises", NPCName: "Elder", CurrentProgress: 0, TargetProgress: 1},
		{ID: "npc_5", Title: "Investigate Strange Noises", NPCName: "Elder", CurrentProgress: 0, TargetProgress: 1},
		{ID: "npc_5", Title: "Investigate Strange Noises", NPCName: "Elder", CurrentProgress: 0, TargetProgress: 1},
		{ID: "npc_5", Title: "Investigate Strange Noises", NPCName: "Elder", CurrentProgress: 0, TargetProgress: 1},
	}
}

// Toggle changes whether the quest panel is open.
func (qm *QuestManager) Toggle() {
	qm.IsOpen = !qm.IsOpen
}

// AddQuest appends quest to the manager's quest list.
func (qm *QuestManager) AddQuest(quest *Quest) {
	qm.Quests = append(qm.Quests, quest)
}

// AcceptQuest adds the quest offered by npc unless its ID is empty or already present.
// npc provides the quest ID and display name; the method reports whether a quest was added.
func (qm *QuestManager) AcceptQuest(npc npc.NPCConfig) bool {
	if npc.QuestID == "" {
		return false
	}
	if qm.HasQuest(npc.QuestID) {
		return false
	}

	qm.AddQuest(&Quest{
		ID:              npc.QuestID,
		Title:           fmt.Sprintf("Quest from %s", npc.Name),
		NPCName:         npc.Name,
		TargetProgress:  1,
		CurrentProgress: 0,
	})
	return true
}

// HasQuest reports whether a quest with questID is present in the manager.
func (qm *QuestManager) HasQuest(questID string) bool {
	for _, quest := range qm.Quests {
		if quest.ID == questID {
			return true
		}
	}
	return false
}
