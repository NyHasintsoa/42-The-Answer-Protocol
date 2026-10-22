package service

import (
	"fmt"
	"math/rand"
	"tap-gui/src/model/npc"
	"time"
)

type AttackManager struct {
	IsOpen            bool
	PlayerHP          int
	PlayerMaxHP       int
	NPC               npc.NPCConfig
	NPCHP             int
	NPCMaxHP          int
	CombatLog         []string
	IsCombatOver      bool
	OnPlayerRespawn   func()
	OnMonsterDefeated func(npc npc.NPCConfig)
	OnOpenInventory   func()
	OnFlee            func()
	OnLog             func(string)
	respawnPending    bool
	respawnRemaining  float32
}

func NewAttackManager() *AttackManager {
	return &AttackManager{
		IsOpen:      false,
		PlayerHP:    100,
		PlayerMaxHP: 100,
		CombatLog:   make([]string, 0),
	}
}

func (am *AttackManager) StartCombat(npc npc.NPCConfig) {
	am.NPC = npc
	am.IsOpen = true
	am.IsCombatOver = false
	am.CombatLog = make([]string, 0)

	if npc.HP <= 0 {
		am.NPCHP = npc.MaxHP
	} else {
		am.NPCHP = npc.HP
	}
	am.NPCMaxHP = npc.MaxHP
	if am.PlayerHP <= 0 {
		am.PlayerHP = am.PlayerMaxHP
	}

	am.AddLog(fmt.Sprintf("Combat started with %s!", npc.Name))
}

func (am *AttackManager) AddLog(msg string) {
	am.CombatLog = append(am.CombatLog, msg)
	if len(am.CombatLog) > 10 {
		am.CombatLog = am.CombatLog[1:]
	}
	if am.OnLog != nil {
		am.OnLog(msg)
	}
}

func (am *AttackManager) Attack() {
	if am.IsCombatOver || !am.IsOpen {
		return
	}

	playerDmg := 15 + rand.Intn(11)
	am.NPCHP -= playerDmg
	if am.NPCHP < 0 {
		am.NPCHP = 0
	}
	am.AddLog(fmt.Sprintf("You dealt %d damage to %s.", playerDmg, am.NPC.Name))

	if am.NPCHP <= 0 {
		am.IsCombatOver = true

		successMsg := fmt.Sprintf("SUCCESS: %s has been defeated!", am.NPC.Name)
		am.AddLog(successMsg)

		if am.OnMonsterDefeated != nil {
			am.NPC.HP = 0
			am.NPC.CanBeAttacked = false
			am.OnMonsterDefeated(am.NPC)
		}

		go func() {
			time.Sleep(3 * time.Second)
			am.Close()
		}()
		return
	}

	monsterDmg := 8 + rand.Intn(10)
	am.PlayerHP -= monsterDmg
	if am.PlayerHP < 0 {
		am.PlayerHP = 0
	}
	am.AddLog(fmt.Sprintf("%s counter-attacked for %d damage.", am.NPC.Name, monsterDmg))

	if am.PlayerHP <= 0 {
		am.IsCombatOver = true
		am.AddLog("You were defeated in combat! Respawning in 3 seconds...")
		am.scheduleRespawn()
	}
}

func (am *AttackManager) Status() {
	if !am.IsOpen {
		return
	}
	am.AddLog(fmt.Sprintf("STATUS: Player HP: %d/%d | %s HP: %d/%d", am.PlayerHP, am.PlayerMaxHP, am.NPC.Name, am.NPCHP, am.NPCMaxHP))
}

func (am *AttackManager) Flee() {
	if am.IsCombatOver || !am.IsOpen {
		return
	}

	if rand.Float32() < 0.7 {
		am.AddLog("You successfully fled from combat!")
		am.IsCombatOver = true
		am.IsOpen = false
		if am.OnFlee != nil {
			am.OnFlee()
		}
	} else {
		am.AddLog("Flee failed!")
		monsterDmg := 5 + rand.Intn(8)
		am.PlayerHP -= monsterDmg
		if am.PlayerHP < 0 {
			am.PlayerHP = 0
		}
		am.AddLog(fmt.Sprintf("%s hit you for %d damage while fleeing.", am.NPC.Name, monsterDmg))
		if am.PlayerHP <= 0 {
			am.IsCombatOver = true
			am.AddLog("You were defeated! Respawning in 3 seconds...")
			am.scheduleRespawn()
		}
	}
}

func (am *AttackManager) UseInventory() {
	if am.IsCombatOver || !am.IsOpen {
		return
	}
	am.AddLog("Opened inventory during combat.")
	if am.OnOpenInventory != nil {
		am.OnOpenInventory()
	}
}

func (am *AttackManager) RespawnPlayer() {
	am.respawnPending = false
	am.respawnRemaining = 0
	am.PlayerHP = 50
	am.AddLog("Respawned at safe location with 50 HP.")
	if am.OnPlayerRespawn != nil {
		am.OnPlayerRespawn()
	}
	am.Close()
}

func (am *AttackManager) Update(dt float32) {
	if !am.respawnPending {
		return
	}
	am.respawnRemaining -= dt
	if am.respawnRemaining <= 0 {
		am.RespawnPlayer()
	}
}

func (am *AttackManager) scheduleRespawn() {
	am.respawnPending = true
	am.respawnRemaining = 3
}

func (am *AttackManager) Close() {
	am.IsOpen = false
}
