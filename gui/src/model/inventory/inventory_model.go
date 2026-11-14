// Package inventory defines the data model for player inventory items.
package inventory

// InventoryCategory identifies an inventory item's gameplay category.
type InventoryCategory string

const (
	// Attack identifies an item category that affects attack power.
	Attack InventoryCategory = "attack"
	// Heal identifies an item category that restores health.
	Heal InventoryCategory = "heal"
)

// InventoryItemType identifies an item and its matching inventory asset.
type InventoryItemType string

const (
	// ItemBookDagger is the inventory type for the dagger mastery book.
	ItemBookDagger InventoryItemType = "book_dagger"
	// ItemBookSword is the inventory type for the sword combat manual.
	ItemBookSword InventoryItemType = "book_sword"
	// ItemBossHorn is the inventory type for a dragon horn.
	ItemBossHorn InventoryItemType = "boss_horn"
	// ItemDagger is the inventory type for a dagger.
	ItemDagger InventoryItemType = "dagger"
	// ItemMedals is the inventory type for the mayor's medal.
	ItemMedals InventoryItemType = "medals"
	// ItemPotionAttackGM is the inventory type for a grand attack potion.
	ItemPotionAttackGM InventoryItemType = "potion_attack_GM"
	// ItemPotionAttackPM is the inventory type for a petite attack potion.
	ItemPotionAttackPM InventoryItemType = "potion_attack_PM"
	// ItemPotionHealthGM is the inventory type for a grand health potion.
	ItemPotionHealthGM InventoryItemType = "potion_health_GM"
	// ItemPotionHealthPM is the inventory type for a petite health potion.
	ItemPotionHealthPM InventoryItemType = "potion_health_PM"
	// ItemSword is the inventory type for a sword.
	ItemSword InventoryItemType = "sword"
)

// Filename returns the asset filename associated with item type e.
func (e InventoryItemType) Filename() string {
	switch e {
	case ItemBookDagger:
		return "book_dagger.png"
	case ItemBookSword:
		return "book_sword.png"
	case ItemBossHorn:
		return "boss_horn.png"
	case ItemDagger:
		return "dagger.png"
	case ItemMedals:
		return "medals.png"
	case ItemPotionAttackGM:
		return "potion_attack_GM.png"
	case ItemPotionAttackPM:
		return "potion_attack_PM.png"
	case ItemPotionHealthGM:
		return "potion_health_GM.png"
	case ItemPotionHealthPM:
		return "potion_health_PM.png"
	case ItemSword:
		return "sword.png"
	default:
		return string(e) + ".png"
	}
}

// InventoryItem contains the display and quantity data for an inventory entry.
type InventoryItem struct {
	ID       string
	Name     string
	Type     InventoryItemType
	Category InventoryCategory
	Cost     int
	Count    int
}
