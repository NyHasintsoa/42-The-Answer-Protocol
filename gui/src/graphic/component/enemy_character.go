package component

import (
	"tap-gui/src/utils"
)

type EnemyCharacter struct {
	NpcCharacter
}

func NewEnemyCharacter(resolver *utils.PathResolver, relativeDir string, x, y, scale float32) (*EnemyCharacter, error) {
	enemy := &EnemyCharacter{
		NpcCharacter: NewNpcCharacter("Enemy", resolver, relativeDir, scale + 0.06),
	}
	enemy.SetPositionOffset(-50, -50)
	actions := []string{"Front_Idle"}
	for _, act := range actions {
		if err := enemy.LoadAnim(act); err != nil {
			return nil, err
		}
	}
	enemy.SetAction("Front_Idle")
	return enemy, nil
}
