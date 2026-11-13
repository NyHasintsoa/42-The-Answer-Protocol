package model

type QuestStatus string

const (
	QuestActive    QuestStatus = "active"
	QuestCompleted QuestStatus = "completed"
)

type PlayerQuest struct {
	QuestID  string      `json:"quest_id"`
	Status   QuestStatus `json:"status"`
	Progress string      `json:"progress"`
}

type Player struct {
	Username     string                  `json:"username"`
	CurrentRoom  string                  `json:"current_room"`
	HP           int                     `json:"hp"`
	MaxHP        int                     `json:"max_hp"`
	AttackBonus  int                     `json:"attack_bonus"`
	InCombatWith string                  `json:"in_combat_with,omitempty"`
	Inventory    []string                `json:"inventory"`
	Quests       map[string]*PlayerQuest `json:"quests"`
	GroupID      string                  `json:"group_id,omitempty"`
	IsConnected  bool                    `json:"-"`
}

type Group struct {
	ID      string          `json:"id"`
	Leader  string          `json:"leader"`
	Members map[string]bool `json:"members"`
	Invites map[string]bool `json:"invites"`
}
