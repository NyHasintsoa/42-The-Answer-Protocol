package inventory

type InventoryCategory string

const (
	Attack InventoryCategory = "attack"
	Heal   InventoryCategory = "heal"
)

type InventoryItemType string

const (
	ItemBookDagger     InventoryItemType = "book_dagger"
	ItemBookSword      InventoryItemType = "book_sword"
	ItemBossHorn       InventoryItemType = "boss_horn"
	ItemDagger         InventoryItemType = "dagger"
	ItemMedals         InventoryItemType = "medals"
	ItemPotionAttackGM InventoryItemType = "potion_attack_GM"
	ItemPotionAttackPM InventoryItemType = "potion_attack_PM"
	ItemPotionHealthGM InventoryItemType = "potion_health_GM"
	ItemPotionHealthPM InventoryItemType = "potion_health_PM"
	ItemSword          InventoryItemType = "sword"
)

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

type InventoryItem struct {
	ID       string
	Name     string
	Type     InventoryItemType
	Category InventoryCategory
	Cost     int
	Count    int
}
