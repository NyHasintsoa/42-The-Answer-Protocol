package service

import (
	"fmt"
	"tap-gui/src/model/npc"
)

type AttackManager struct {
	IsOpen       bool
	PlayerHP     int
	PlayerMaxHP  int
	NPC          npc.NPCConfig
	NPCHP        int
	NPCMaxHP     int
	CombatLog    []string
	IsCombatOver bool
	OnLog        func(string)
}

// NewAttackManager creates an empty combat view populated by server responses.
func NewAttackManager() *AttackManager {
	return &AttackManager{
		PlayerHP:    100,
		PlayerMaxHP: 100,
		CombatLog:   make([]string, 0),
	}
}

// StartCombat opens the combat view for a server-defined opponent.
func (am *AttackManager) StartCombat(opponent npc.NPCConfig) {
	am.NPC = opponent
	am.IsOpen = true
	am.IsCombatOver = false
	am.CombatLog = nil
	am.NPCHP = opponent.HP
	if am.NPCHP <= 0 {
		am.NPCHP = opponent.MaxHP
	}
	am.NPCMaxHP = opponent.MaxHP
	am.AddLog(fmt.Sprintf("Combat started with %s.", opponent.Name))
}

// AddLog appends a server response to the combat history.
func (am *AttackManager) AddLog(message string) {
	am.CombatLog = append(am.CombatLog, message)
	if len(am.CombatLog) > 10 {
		am.CombatLog = am.CombatLog[len(am.CombatLog)-10:]
	}
	if am.OnLog != nil {
		am.OnLog(message)
	}
}

// Close hides the combat panel.
func (am *AttackManager) Close() {
	am.IsOpen = false
}
