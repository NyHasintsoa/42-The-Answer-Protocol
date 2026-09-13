package component

import (
	"tap-gui/src/utils"
)

type PersonCharacter struct {
	NpcCharacter
}

func NewPersonCharacter(resolver *utils.PathResolver, relativeDir string, x, y, scale float32) (*PersonCharacter, error) {
	person := &PersonCharacter{
		NpcCharacter: NewNpcCharacter("Person", resolver, relativeDir, x, y, scale),
	}
	actions := []string{"Communication", "Idle", "Greeting", "Idle_Blinking"}
	for _, act := range actions {
		if err := person.LoadAnim(act); err != nil {
			return nil, err
		}
	}
	return person, nil
}