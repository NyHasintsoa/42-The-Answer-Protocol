package service

import "fmt"

type ActionResponse struct {
	Index     int
	Text      string
	ActionTag string
	OnSelect  func()
}

type ActionManager struct {
	IsOpen        bool
	NPC           NPCConfig
	Responses     []ActionResponse
	CurrentText   string
	DialogueIndex int
	QuestAccepted bool
	HasQuest      func(NPCConfig) bool
	OnAcceptQuest func(NPCConfig) bool
}

func NewActionManager() *ActionManager {
	return &ActionManager{
		IsOpen:    false,
		Responses: []ActionResponse{},
	}
}

func (dm *ActionManager) OpenAction(npc NPCConfig) {
	dm.NPC = npc
	dm.IsOpen = true
	dm.DialogueIndex = 0
	dm.QuestAccepted = npc.QuestID != "" && dm.HasQuest != nil && dm.HasQuest(npc)
	if len(npc.Dialogue) > 0 {
		dm.CurrentText = npc.Dialogue[0]
	} else {
		dm.CurrentText = "What would you like to do?"
	}
	dm.refreshResponses()
}

func (dm *ActionManager) refreshResponses() {
	dm.Responses = make([]ActionResponse, 0, 3)
	addResponse := func(text, tag string, onSelect func()) {
		dm.Responses = append(dm.Responses, ActionResponse{
			Index:     len(dm.Responses) + 1,
			Text:      text,
			ActionTag: tag,
			OnSelect:  onSelect,
		})
	}

	if len(dm.NPC.Dialogue) > 0 {
		addResponse("Continue conversation", "TALK", dm.continueDialogue)
	}
	if dm.NPC.QuestID != "" && !dm.QuestAccepted {
		addResponse("Accept this quest", "QUEST", dm.acceptQuest)
	}
	addResponse("Leave", "CLOSE", dm.Close)
}

func (dm *ActionManager) continueDialogue() {
	if len(dm.NPC.Dialogue) == 0 {
		return
	}
	dm.DialogueIndex = (dm.DialogueIndex + 1) % len(dm.NPC.Dialogue)
	dm.CurrentText = dm.NPC.Dialogue[dm.DialogueIndex]
}

func (dm *ActionManager) acceptQuest() {
	if dm.OnAcceptQuest != nil && dm.OnAcceptQuest(dm.NPC) {
		dm.QuestAccepted = true
		dm.CurrentText = fmt.Sprintf("Quest accepted: %s", dm.NPC.Name)
	} else {
		dm.CurrentText = "This quest is already in your quest log."
		dm.QuestAccepted = true
	}
	dm.refreshResponses()
}

func (dm *ActionManager) Close() {
	dm.IsOpen = false
}

func (dm *ActionManager) Toggle() {
	dm.IsOpen = !dm.IsOpen
}