package character

import (
	"tap-gui/src/utils"
)

type PersonCharacter struct {
	NpcCharacter
}

// NewPersonCharacter creates a person preview and loads its communication and idle animations.
// resolver loads character assets; relativeDir identifies the asset directory;
// scale controls the rendered size.
func NewPersonCharacter(resolver *utils.PathResolver, relativeDir string, scale float32) (*PersonCharacter, error) {
	person := &PersonCharacter{
		NpcCharacter: NewNpcCharacter("Person", resolver, relativeDir, scale),
	}
	actions := []string{"Communication", "Idle", "Greeting", "Idle_Blinking"}
	for _, act := range actions {
		if err := person.LoadAnim(act); err != nil {
			return nil, err
		}
	}
	return person, nil
}
