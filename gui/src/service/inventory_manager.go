package service

import (
	"tap-gui/src/model/inventory"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// InventoryManager stores server-reported inventory and controls its UI state.
type InventoryManager struct {
	Items             []inventory.InventoryItem
	IsOpen            bool
	CurrentPage       int
	ItemsPerPage      int
	SelectedItemIndex int
	IsModalOpen       bool
}

// NewInventoryManager creates an empty inventory populated by server responses.
func NewInventoryManager() *InventoryManager {
	return &InventoryManager{
		IsOpen:            false,
		CurrentPage:       0,
		ItemsPerPage:      9,
		SelectedItemIndex: -1,
		IsModalOpen:       false,
		Items:             []inventory.InventoryItem{},
	}
}

// SetItems replaces inventory entries with the item records returned by the server.
func (im *InventoryManager) SetItems(items []inventory.InventoryItem) {
	im.Items = append([]inventory.InventoryItem(nil), items...)
	im.CurrentPage = 0
	im.SelectedItemIndex = -1
	im.IsModalOpen = false
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
