package component

import (
	"tap-gui/src/utils"
)

type EnemyCharacter struct {
	NpcCharacter
}

func NewEnemyCharacter(resolver *utils.PathResolver, relativeDir string, x, y, scale float32) (*EnemyCharacter, error) {
	enemy := &EnemyCharacter{
		NpcCharacter: NewNpcCharacter("Enemy", resolver, relativeDir, x - 35, y - 15, scale + 0.06),
	}
	actions := []string{"Front_Idle"}
	for _, act := range actions {
		if err := enemy.LoadAnim(act); err != nil {
			return nil, err
		}
	}
	enemy.SetAction("Front_Idle")
	return enemy, nil
}
