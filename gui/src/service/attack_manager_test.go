package service

import "testing"

func TestPlayerRespawnWaitsThreeSeconds(t *testing.T) {
	manager := NewAttackManager()
	manager.StartCombat(NPCConfig{Name: "enemy"})
	manager.PlayerHP = 1

	respawnCount := 0
	manager.OnPlayerRespawn = func() {
		respawnCount++
	}

	manager.Attack()
	if respawnCount != 0 {
		t.Fatal("player respawned before the delay elapsed")
	}

	manager.Update(2)
	if respawnCount != 0 {
		t.Fatal("player respawned before three seconds")
	}

	manager.Update(1)
	if respawnCount != 1 {
		t.Fatalf("expected one respawn after three seconds, got %d", respawnCount)
	}
	if manager.PlayerHP != 50 {
		t.Fatalf("expected respawn HP 50, got %d", manager.PlayerHP)
	}
	if manager.IsOpen {
		t.Fatal("combat remained open after respawn")
	}
}
