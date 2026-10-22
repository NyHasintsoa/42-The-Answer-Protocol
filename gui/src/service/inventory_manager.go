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

func (im *InventoryManager) GetTotalPages() int {
	if len(im.Items) == 0 {
		return 1
	}
	return (len(im.Items) + im.ItemsPerPage - 1) / im.ItemsPerPage
}

func (im *InventoryManager) NextPage() {
	if im.CurrentPage < im.GetTotalPages()-1 {
		im.CurrentPage++
	}
}

func (im *InventoryManager) PrevPage() {
	if im.CurrentPage > 0 {
		im.CurrentPage--
	}
}

func (im *InventoryManager) Toggle() {
	im.IsOpen = !im.IsOpen
	if !im.IsOpen {
		im.IsModalOpen = false
		im.InCombatMode = false
	}
}

func (im *InventoryManager) SelectItem(index int) {
	if index >= 0 && index < len(im.Items) {
		im.SelectedItemIndex = index
		im.IsModalOpen = true
	}
}

func (im *InventoryManager) CloseModal() {
	im.IsModalOpen = false
	im.SelectedItemIndex = -1
}

func (im *InventoryManager) UseItem(index int) {
	if index < 0 || index >= len(im.Items) {
		return
	}

	item := im.Items[index]
	if im.OnUseItemInCombat != nil && im.InCombatMode {
		im.OnUseItemInCombat(item)
	}

	im.Items[index].Count--
	if im.Items[index].Count <= 0 {
		im.Items = append(im.Items[:index], im.Items[index+1:]...)
	}

	im.CloseModal()
	if im.InCombatMode {
		im.IsOpen = false
		im.InCombatMode = false
	}
}

func (im *InventoryManager) TakeItem(index int) {
	if index < 0 || index >= len(im.Items) {
		return
	}
	im.Items[index].Count++
}

func (im *InventoryManager) DropItem(index int) (inventory.InventoryItem, bool) {
	if index < 0 || index >= len(im.Items) {
		return inventory.InventoryItem{}, false
	}

	droppedItem := im.Items[index]
	im.Items[index].Count--
	if im.Items[index].Count <= 0 {
		im.Items = append(im.Items[:index], im.Items[index+1:]...)
		im.CloseModal()

		if im.CurrentPage >= im.GetTotalPages() && im.CurrentPage > 0 {
			im.CurrentPage--
		}
	}
	return droppedItem, true
}

func (im *InventoryManager) Update(winWidth, winHeight int32) {
	if rl.IsKeyPressed(rl.KeyI) {
		im.Toggle()
	}

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
