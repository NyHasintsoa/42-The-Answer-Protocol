package service

import (
	"tap-gui/src/model/inventory"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type InventoryManager struct {
	Items             []inventory.InventoryItem
	IsOpen            bool
	CurrentPage       int
	ItemsPerPage      int
	SelectedItemIndex int
	IsModalOpen       bool
	InCombatMode      bool
	OnUseItemInCombat func(item inventory.InventoryItem)
}

// NewInventoryManager creates an inventory manager populated with the default sample items.
func NewInventoryManager() *InventoryManager {
	return &InventoryManager{
		IsOpen:            false,
		CurrentPage:       0,
		ItemsPerPage:      9,
		SelectedItemIndex: -1,
		IsModalOpen:       false,
		InCombatMode:      false,
		Items: []inventory.InventoryItem{
			{ID: "book_dagger", Name: "Dagger Book", Type: inventory.ItemBookDagger, Count: 2, Cost: 25, Category: inventory.Attack},
			{ID: "book_sword", Name: "Sword Book", Type: inventory.ItemBookSword, Count: 1, Cost: 25, Category: inventory.Attack},
			{ID: "boss_horn", Name: "Boss Horn", Type: inventory.ItemBossHorn, Count: 3, Cost: 25, Category: inventory.Attack},
			{ID: "dagger", Name: "Dagger", Type: inventory.ItemDagger, Count: 5, Cost: 25, Category: inventory.Attack},
			{ID: "medals", Name: "Medals", Type: inventory.ItemMedals, Count: 10, Cost: 25, Category: inventory.Attack},
			{ID: "sword", Name: "Sword", Type: inventory.ItemSword, Count: 1, Cost: 25, Category: inventory.Attack},
			{ID: "potion_attack_GM", Name: "Attack Potion GM", Type: inventory.ItemPotionAttackGM, Count: 4, Cost: 25, Category: inventory.Heal},
			{ID: "potion_attack_PM", Name: "Attack Potion PM", Type: inventory.ItemPotionAttackPM, Count: 8, Cost: 25, Category: inventory.Heal},
			{ID: "potion_health_GM", Name: "Health Potion GM", Type: inventory.ItemPotionHealthGM, Count: 6, Cost: 25, Category: inventory.Heal},
			{ID: "potion_health_PM", Name: "Health Potion PM", Type: inventory.ItemPotionHealthPM, Count: 12, Cost: 25, Category: inventory.Heal},
		},
	}
}

// GetTotalPages returns the number of pages required to display all items.
func (im *InventoryManager) GetTotalPages() int {
	if len(im.Items) == 0 {
		return 1
	}
	return (len(im.Items) + im.ItemsPerPage - 1) / im.ItemsPerPage
}

// NextPage advances the current page when another page is available.
func (im *InventoryManager) NextPage() {
	if im.CurrentPage < im.GetTotalPages()-1 {
		im.CurrentPage++
	}
}

// PrevPage moves to the previous page when the current page is not the first.
func (im *InventoryManager) PrevPage() {
	if im.CurrentPage > 0 {
		im.CurrentPage--
	}
}

// Toggle opens or closes the inventory and resets modal state when closing it.
func (im *InventoryManager) Toggle() {
	im.IsOpen = !im.IsOpen
	if !im.IsOpen {
		im.IsModalOpen = false
		im.InCombatMode = false
	}
}

// SelectItem selects the item at index and opens its detail modal when the index is valid.
func (im *InventoryManager) SelectItem(index int) {
	if index >= 0 && index < len(im.Items) {
		im.SelectedItemIndex = index
		im.IsModalOpen = true
	}
}

// CloseModal closes the item detail modal and clears the selection.
func (im *InventoryManager) CloseModal() {
	im.IsModalOpen = false
	im.SelectedItemIndex = -1
}

// AddItem adds item to the inventory, merging it into an existing stack by ID.
func (im *InventoryManager) AddItem(item inventory.InventoryItem) bool {
	if item.ID == "" {
		return false
	}
	if item.Count <= 0 {
		item.Count = 1
	}
	for index := range im.Items {
		if im.Items[index].ID == item.ID {
			im.Items[index].Count += item.Count
			return true
		}
	}
	if item.Type == "" {
		item.Type = inventory.InventoryItemType(item.ID)
	}
	im.Items = append(im.Items, item)
	return true
}

// UseItem consumes one item at index and invokes the combat callback when applicable.
func (im *InventoryManager) UseItem(index int) {
	if index < 0 || index >= len(im.Items) {
		return
	}

	item := im.Items[index]
	if im.OnUseItemInCombat != nil && im.InCombatMode {
		im.OnUseItemInCombat(item)
	}

	im.removeOne(index)
	im.CloseModal()
	if im.InCombatMode {
		im.IsOpen = false
		im.InCombatMode = false
	}
}

// TakeItem consumes and returns one item at index.
func (im *InventoryManager) TakeItem(index int) (inventory.InventoryItem, bool) {
	return im.removeOne(index)
}

func (im *InventoryManager) removeOne(index int) (inventory.InventoryItem, bool) {
	if index < 0 || index >= len(im.Items) {
		return inventory.InventoryItem{}, false
	}

	item := im.Items[index]
	im.Items[index].Count--
	if im.Items[index].Count <= 0 {
		im.Items = append(im.Items[:index], im.Items[index+1:]...)
		im.CloseModal()
		if im.CurrentPage >= im.GetTotalPages() && im.CurrentPage > 0 {
			im.CurrentPage--
		}
	}
	return item, true
}

// DropItem removes one unit of the item at index and returns the dropped item.
// index identifies the inventory entry; the boolean is false if the index is invalid.
func (im *InventoryManager) DropItem(index int) (inventory.InventoryItem, bool) {
	if index < 0 || index >= len(im.Items) {
		return inventory.InventoryItem{}, false
	}

	return im.removeOne(index)
}

// Update handles inventory keyboard and mouse input for a window of the given size.
// winWidth and winHeight are the window dimensions in pixels.
func (im *InventoryManager) Update(winWidth, winHeight int32) {
	mousePos := rl.GetMousePosition()
	mouseClicked := rl.IsMouseButtonPressed(rl.MouseLeftButton)

	btnRect := rl.NewRectangle(20, float32(winHeight-130), 110, 40)
	if mouseClicked && rl.CheckCollisionPointRec(mousePos, btnRect) {
		im.Toggle()
		return
	}

	if !im.IsOpen {
		return
	}

	panelW := float32(500)
	panelH := float32(520)
	panelX := (float32(winWidth) - panelW) / 2
	panelY := (float32(winHeight) - panelH) / 2

	if im.IsModalOpen {
		return
	}

	closeBtn := rl.NewRectangle(panelX+panelW-42, panelY+12, 30, 30)
	if mouseClicked && rl.CheckCollisionPointRec(mousePos, closeBtn) {
		im.IsOpen = false
		im.InCombatMode = false
		return
	}

	prevBtn := rl.NewRectangle(panelX+14, panelY+panelH/2-25, 36, 50)
	if mouseClicked && rl.CheckCollisionPointRec(mousePos, prevBtn) {
		im.PrevPage()
		return
	}

	nextBtn := rl.NewRectangle(panelX+panelW-50, panelY+panelH/2-25, 36, 50)
	if mouseClicked && rl.CheckCollisionPointRec(mousePos, nextBtn) {
		im.NextPage()
		return
	}
}
