package model

import "time"

type InteractionType string

const (
	InteractionChat   InteractionType = "CHAT"
	InteractionMove   InteractionType = "MOVE"
	InteractionItem   InteractionType = "ITEM"
	InteractionCombat InteractionType = "COMBAT"
	InteractionQuest  InteractionType = "QUEST"
	InteractionSystem InteractionType = "SYSTEM"
)

type InteractionRecord struct {
	ID        string          `json:"id"`
	Timestamp time.Time       `json:"timestamp"`
	Level     string          `json:"level"` // INFO, WARN, ERROR
	Type      InteractionType `json:"type"`
	Event     string          `json:"event,omitempty"`
	Actor     string          `json:"actor"`
	Remote    string          `json:"remote,omitempty"`
	Command   string          `json:"command,omitempty"`
	Response  string          `json:"response,omitempty"`
	Scope     string          `json:"scope,omitempty"`  // GLOBAL, ROOM, GROUP
	Target    string          `json:"target,omitempty"` // Room ID, NPC ID, Item ID
	Payload   string          `json:"payload"`
}

type InteractionDatabase struct {
	UpdatedAt    time.Time           `json:"updated_at"`
	Players      map[string]*Player  `json:"players"`
	Interactions []InteractionRecord `json:"interactions"`
}

// LookResponse matches RFC 42TAP Section 5.1.2
type LookResponse struct {
	Room    RoomSummary `json:"room"`
	Players []string    `json:"players"`
	Items   []string    `json:"items"`
	NPCs    []string    `json:"npcs"`
}

type RoomSummary struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Exits       map[string]string `json:"exits"`
}
