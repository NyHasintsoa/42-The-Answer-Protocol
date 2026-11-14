package service

import "tap-gui/src/model/npc"

type ActionResponse struct {
	Index     int
	Text      string
	ActionTag string
}

type ActionManager struct {
	IsOpen        bool
	NPC           npc.NPCConfig
	Responses     []ActionResponse
	CurrentText   string
	QuestAccepted bool
	HasQuest      func(npc.NPCConfig) bool
}

// NewActionManager creates a closed action manager with no available responses.
func NewActionManager() *ActionManager {
	return &ActionManager{
		IsOpen:    false,
		Responses: []ActionResponse{},
	}
}

// OpenAction opens the dialogue and available responses for npc.
// npc supplies the dialogue lines, quest, and name shown in the action panel.
func (dm *ActionManager) OpenAction(npc npc.NPCConfig) {
	dm.NPC = npc
	dm.IsOpen = true
	dm.QuestAccepted = npc.QuestID != "" && dm.HasQuest != nil && dm.HasQuest(npc)
	if len(npc.Dialogue) > 0 {
		dm.CurrentText = npc.Dialogue[0]
	} else {
		dm.CurrentText = "What would you like to do?"
	}
	dm.refreshResponses()
}

// refreshResponses rebuilds the response list for the current NPC and dialogue state.
func (dm *ActionManager) refreshResponses() {
	dm.Responses = make([]ActionResponse, 0, 3)
	addResponse := func(text, tag string) {
		dm.Responses = append(dm.Responses, ActionResponse{
			Index:     len(dm.Responses) + 1,
			Text:      text,
			ActionTag: tag,
		})
	}

	if len(dm.NPC.Dialogue) > 0 {
		addResponse("Continue conversation", "TALK")
	}
	if dm.NPC.QuestID != "" && !dm.QuestAccepted {
		addResponse("Accept this quest", "QUEST")
	}
	addResponse("Leave", "CLOSE")
}

// Close hides the action panel.
func (dm *ActionManager) Close() {
	dm.IsOpen = false
}

// Toggle changes whether the action panel is open.
func (dm *ActionManager) Toggle() {
	dm.IsOpen = !dm.IsOpen
}

// SetQuestAccepted updates the quest action based on a server response.
func (dm *ActionManager) SetQuestAccepted(accepted bool) {
	dm.QuestAccepted = accepted
	dm.refreshResponses()
}
