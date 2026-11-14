package service

// Quest stores a player's displayed quest state.
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

// QuestManager stores quest state and controls the quest panel.
type QuestManager struct {
	IsOpen bool
	Quests []*Quest
}

// NewQuestManager creates an empty quest manager populated by server responses.
func NewQuestManager() *QuestManager {
	return &QuestManager{
		IsOpen: false,
		Quests: []*Quest{},
	}
}

// Toggle changes whether the quest panel is open.
func (qm *QuestManager) Toggle() {
	qm.IsOpen = !qm.IsOpen
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

// SetQuests replaces the displayed quest list with the state returned by the server.
func (qm *QuestManager) SetQuests(quests []*Quest) {
	qm.Quests = append([]*Quest(nil), quests...)
}
