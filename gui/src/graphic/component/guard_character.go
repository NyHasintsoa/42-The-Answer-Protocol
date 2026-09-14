package component

import (
	"tap-gui/src/utils"
)

type GuardCharacter struct {
	NpcCharacter
}

func NewGuardCharacter(resolver *utils.PathResolver, relativeDir string, x, y, scale float32) (*GuardCharacter, error) {
	guard := &GuardCharacter{
		NpcCharacter: NewNpcCharacter("Guard", resolver, relativeDir, x - 30, y - 25, scale),
	}
	actions := []string{"Communication", "Idle", "Idle_Blinking"}
	for _, act := range actions {
		if err := guard.LoadAnim(act); err != nil {
			return nil, err
		}
	}
	return guard, nil
}