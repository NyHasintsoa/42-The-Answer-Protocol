package component

import (
	"tap-gui/src/utils"
)

type SellerCharacter struct {
	NpcCharacter
}

func NewSellerCharacter(resolver *utils.PathResolver, relativeDir string, x, y, scale float32) (*SellerCharacter, error) {
	seller := &SellerCharacter{
		NpcCharacter: NewNpcCharacter("Seller", resolver, relativeDir, x, y, scale),
	}
	actions := []string{"Chagrin", "Communication", "Greeting", "Idle", "Idle_Blinking", "Joy"}
	for _, act := range actions {
		if err := seller.LoadAnim(act); err != nil {
			return nil, err
		}
	}
	return seller, nil
}