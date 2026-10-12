package character

import (
	"tap-gui/src/utils"
)

type MerchantCharacter struct {
	NpcCharacter
}

func NewMerchantCharacter(resolver *utils.PathResolver, relativeDir string, scale float32) (*MerchantCharacter, error) {
	merchant := &MerchantCharacter{
		NpcCharacter: NewNpcCharacter("Merchant", resolver, relativeDir, scale),
	}
	actions := []string{"Chagrin", "Communication", "Greeting", "Idle", "Idle_Blinking", "Joy"}
	for _, act := range actions {
		if err := merchant.LoadAnim(act); err != nil {
			return nil, err
		}
	}
	return merchant, nil
}
