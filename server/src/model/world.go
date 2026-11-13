package model

type WorldConfig struct {
	World  WorldData         `json:"world"`
	NPCs   map[string]*NPC   `json:"npcs"`
	Items  map[string]*Item  `json:"items"`
	Quests map[string]*Quest `json:"quests"`
}

type WorldData struct {
	Rooms map[string]*Room `json:"rooms"`
}

type Room struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Exits       map[string]string `json:"exits"`
	Items       []string          `json:"items"`
	NPCs        []string          `json:"npcs"`
}

type NPC struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Type          string   `json:"type"`
	ImageType     string   `json:"image_type"`
	CanBeAttacked bool     `json:"can_be_attacked"`
	HP            int      `json:"hp"`
	MaxHP         int      `json:"max_hp"`
	AttackDamage  int      `json:"attack_damage"`
	QuestID       string   `json:"quest_id"`
	Dialogue      []string `json:"dialogue"`
}

type Item struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Cost        int    `json:"cost"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

type Quest struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Giver        string `json:"giver"`
	Description  string `json:"description"`
	RequiredItem string `json:"required_item"`
	Reward       string `json:"reward"`
}
