package service

type Quest struct {
	ID              string
	Title           string
	NPCName         string
	CurrentProgress int
	TargetProgress  int
}

func (q *Quest) IsCompleted() bool {
	return q.CurrentProgress >= q.TargetProgress
}

type QuestManager struct {
	IsOpen bool
	Quests []*Quest
}

func NewQuestManager() *QuestManager {
	qm := &QuestManager{
		IsOpen: false,
		Quests: []*Quest{},
	}
	qm.initDefaultQuests()
	return qm
}

func (qm *QuestManager) initDefaultQuests() {
	qm.Quests = []*Quest{
		{ID: "npc_1", Title: "Talk to Guard Captain", NPCName: "Guard", CurrentProgress: 1, TargetProgress: 1},
		{ID: "npc_2", Title: "Gather 5 Herbs for Seller", NPCName: "Seller", CurrentProgress: 2, TargetProgress: 5},
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

func (qm *QuestManager) Toggle() {
	qm.IsOpen = !qm.IsOpen
}

func (qm *QuestManager) AddQuest(quest *Quest) {
	qm.Quests = append(qm.Quests, quest)
}