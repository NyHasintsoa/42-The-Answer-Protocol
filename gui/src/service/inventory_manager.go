package service

import (
	"fmt"
	"tap-gui/src/model/enums"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type ItemCategory string

const (
	CategoryWeapon ItemCategory = "Weapon"
	CategoryArmor  ItemCategory = "Armor"
	CategoryPotion ItemCategory = "Potion"
	CategoryKey    ItemCategory = "Key"
	CategoryMisc   ItemCategory = "Misc"
)

type InventoryItem struct {
	ID       string
	Name     string
	Type     enums.InventoryItemType
	Count    int
	Category ItemCategory
	Color    rl.Color
}

type InventoryManager struct {
	Items             []InventoryItem
	IsOpen            bool
	CurrentPage       int
	ItemsPerPage      int
	SelectedItemIndex int
	IsModalOpen       bool
}

func NewInventoryManager() *InventoryManager {
	return &InventoryManager{
		IsOpen:            false,
		CurrentPage:       0,
		ItemsPerPage:      9,
		SelectedItemIndex: -1,
		IsModalOpen:       false,
		Items: []InventoryItem{
			{ID: "book_dagger", Name: "Dagger Book", Type: enums.ItemBookDagger, Count: 2, Category: CategoryMisc, Color: rl.NewColor(180, 130, 80, 255)},
			{ID: "book_sword", Name: "Sword Book", Type: enums.ItemBookSword, Count: 1, Category: CategoryMisc, Color: rl.NewColor(200, 210, 225, 255)},
			{ID: "boss_horn", Name: "Boss Horn", Type: enums.ItemBossHorn, Count: 3, Category: CategoryMisc, Color: rl.NewColor(255, 165, 0, 255)},
			{ID: "dagger", Name: "Dagger", Type: enums.ItemDagger, Count: 5, Category: CategoryWeapon, Color: rl.NewColor(180, 180, 180, 255)},
			{ID: "medals", Name: "Medals", Type: enums.ItemMedals, Count: 10, Category: CategoryMisc, Color: rl.NewColor(245, 200, 50, 255)},
			{ID: "potion_attack_GM", Name: "Attack Potion GM", Type: enums.ItemPotionAttackGM, Count: 4, Category: CategoryPotion, Color: rl.NewColor(230, 70, 50, 255)},
			{ID: "potion_attack_PM", Name: "Attack Potion PM", Type: enums.ItemPotionAttackPM, Count: 8, Category: CategoryPotion, Color: rl.NewColor(240, 120, 90, 255)},
			{ID: "potion_health_GM", Name: "Health Potion GM", Type: enums.ItemPotionHealthGM, Count: 6, Category: CategoryPotion, Color: rl.NewColor(50, 205, 80, 255)},
			{ID: "potion_health_PM", Name: "Health Potion PM", Type: enums.ItemPotionHealthPM, Count: 12, Category: CategoryPotion, Color: rl.NewColor(100, 220, 120, 255)},
			{ID: "sword", Name: "Sword", Type: enums.ItemSword, Count: 1, Category: CategoryWeapon, Color: rl.NewColor(130, 140, 155, 255)},
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

func (im *InventoryManager) TakeItem(index int) {
	if index < 0 || index >= len(im.Items) {
		return
	}
	im.Items[index].Count++
	fmt.Printf("Take: %s\n", im.Items[index].Name)
}

func (im *InventoryManager) DropItem(index int) {
	if index < 0 || index >= len(im.Items) {
		return
	}

	fmt.Printf("Drop: %s\n", im.Items[index].Name)
	im.Items[index].Count--

	if im.Items[index].Count <= 0 {
		im.Items = append(im.Items[:index], im.Items[index+1:]...)
		im.CloseModal()

		if im.CurrentPage >= im.GetTotalPages() && im.CurrentPage > 0 {
			im.CurrentPage--
		}
	}
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