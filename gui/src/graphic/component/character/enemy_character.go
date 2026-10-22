package character

import (
	"tap-gui/src/utils"
)

type EnemyCharacter struct {
	NpcCharacter
}

// NewEnemyCharacter creates an enemy preview and loads its idle animation.
// resolver loads the character assets; relativeDir identifies the asset directory;
// scale controls the rendered size.
func NewEnemyCharacter(resolver *utils.PathResolver, relativeDir string, scale float32) (*EnemyCharacter, error) {
	enemy := &EnemyCharacter{
		NpcCharacter: NewNpcCharacter("Enemy", resolver, relativeDir, scale+0.06),
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
