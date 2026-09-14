package component

import (
	"fmt"

	"tap-gui/src/model/enums"
	"tap-gui/src/service"
	"tap-gui/src/utils"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type InventoryComponent struct {
	Manager  *service.InventoryManager
	Resolver *utils.PathResolver
	textures map[enums.InventoryItemType]rl.Texture2D
}

func NewInventoryComponent(mgr *service.InventoryManager, resolver *utils.PathResolver) *InventoryComponent {
	return &InventoryComponent{
		Manager:  mgr,
		Resolver: resolver,
		textures: make(map[enums.InventoryItemType]rl.Texture2D),
	}
}

func (ic *InventoryComponent) GetTexture(itemType enums.InventoryItemType) rl.Texture2D {
	if tex, exists := ic.textures[itemType]; exists {
		return tex
	}

	var imagePath string
	if ic.Resolver != nil {
		imagePath = ic.Resolver.Resolve("inventory", itemType.Filename())
	} else {
		imagePath = "inventory/" + itemType.Filename()
	}

	tex := ic.Resolver.ImageManager.Load(imagePath)
	ic.textures[itemType] = tex
	return tex
}

func (ic *InventoryComponent) Unload() {
	ic.textures = make(map[enums.InventoryItemType]rl.Texture2D)
}

func (ic *InventoryComponent) Render(winWidth, winHeight int32) {
	if ic.Manager == nil || !ic.Manager.IsOpen {
		return
	}

	mousePos := rl.GetMousePosition()
	mouseClicked := rl.IsMouseButtonPressed(rl.MouseLeftButton)

	if rl.IsKeyPressed(rl.KeyEscape) {
		if ic.Manager.IsModalOpen {
			ic.Manager.CloseModal()
		} else {
			ic.Manager.IsOpen = false
		}
		return
	}

	rl.DrawRectangle(0, 0, winWidth, winHeight, rl.NewColor(0, 0, 0, 160))

	panelW := float32(500)
	panelH := float32(520)
	panelX := (float32(winWidth) - panelW) / 2
	panelY := (float32(winHeight) - panelH) / 2
	panelRect := rl.NewRectangle(panelX, panelY, panelW, panelH)

	rl.DrawRectangleRounded(panelRect, 0.08, 12, rl.NewColor(22, 25, 38, 245))
	rl.DrawRectangleRoundedLinesEx(panelRect, 0.08, 12, 3, rl.NewColor(0, 180, 255, 255))

	headerRect := rl.NewRectangle(panelX, panelY, panelW, 55)
	rl.DrawRectangleRounded(headerRect, 0.15, 8, rl.NewColor(15, 18, 28, 255))
	rl.DrawLineEx(rl.NewVector2(panelX, panelY+55), rl.NewVector2(panelX+panelW, panelY+55), 2, rl.NewColor(0, 180, 255, 255))

	titleText := fmt.Sprintf("INVENTORY (%d / %d)", ic.Manager.CurrentPage+1, ic.Manager.GetTotalPages())
	rl.DrawText(titleText, int32(panelX+24), int32(panelY+17), 22, rl.Gold)

	closeButton := NewButton(panelX+panelW-42, panelY+12, 30, 30, "X", 18)
	closeButton.BgColor = rl.NewColor(180, 50, 50, 255)
	closeButton.HoverBgColor = rl.Red
	closeButton.ClickedColor = closeButton.HoverBgColor
	closeButton.TextColor = rl.White
	closeButton.BorderColor = rl.White
	closeButton.ShadowColor = rl.Blank
	closeButton.BorderWidth = 1
	closeButton.BorderRadius = 0.3
	closeButton.Render()
	if closeButton.IsClicked && !ic.Manager.IsModalOpen {
		ic.Manager.IsOpen = false
	}

	slotSize := float32(115)
	spacing := float32(18)
	gridStartX := panelX + (panelW-3*slotSize-2*spacing)/2
	gridStartY := panelY + 80

	startIndex := ic.Manager.CurrentPage * ic.Manager.ItemsPerPage
	endIndex := startIndex + ic.Manager.ItemsPerPage
	if endIndex > len(ic.Manager.Items) {
		endIndex = len(ic.Manager.Items)
	}

	for i := range 9 {
		row := i / 3
		col := i % 3

		slotX := gridStartX + float32(col)*(slotSize+spacing)
		slotY := gridStartY + float32(row)*(slotSize+spacing)
		slotRect := rl.NewRectangle(slotX, slotY, slotSize, slotSize)

		slotBorderCol := rl.NewColor(70, 80, 110, 255)
		if rl.CheckCollisionPointRec(mousePos, slotRect) && !ic.Manager.IsModalOpen {
			slotBorderCol = rl.Gold
		}

		rl.DrawRectangleRounded(slotRect, 0.15, 8, rl.NewColor(35, 40, 55, 255))
		rl.DrawRectangleRoundedLinesEx(slotRect, 0.15, 8, 2, slotBorderCol)

		itemIndex := startIndex + i
		if itemIndex < endIndex {
			item := ic.Manager.Items[itemIndex]

			if mouseClicked && rl.CheckCollisionPointRec(mousePos, slotRect) && !ic.Manager.IsModalOpen {
				ic.Manager.SelectItem(itemIndex)
			}

			tex := ic.GetTexture(item.Type)
			if tex.ID > 0 {
				srcRec := rl.NewRectangle(0, 0, float32(tex.Width), float32(tex.Height))
				destRec := slotRect
				rl.DrawTexturePro(tex, srcRec, destRec, rl.Vector2{X: 0, Y: 0}, 0, rl.White)
			} else {
				rl.DrawRectangleRounded(slotRect, 0.15, 8, item.Color)
			}

			countStr := fmt.Sprintf("x%d", item.Count)
			badgeW := float32(rl.MeasureText(countStr, 12) + 10)
			badgeRect := rl.NewRectangle(slotX+slotSize-badgeW-4, slotY+slotSize-22, badgeW, 18)
			rl.DrawRectangleRounded(badgeRect, 0.3, 6, rl.NewColor(15, 18, 28, 230))
			rl.DrawRectangleRoundedLinesEx(badgeRect, 0.3, 6, 1, rl.NewColor(0, 180, 255, 255))
			rl.DrawText(countStr, int32(slotX+slotSize-badgeW+2), int32(slotY+slotSize-19), 12, rl.Gold)
		}
	}

	prevButton := NewButton(panelX+14, panelY+panelH/2-25, 36, 50, "<", 28)
	prevButton.BgColor = rl.NewColor(0, 100, 180, 255)
	prevButton.HoverBgColor = rl.NewColor(0, 150, 220, 255)
	prevButton.ClickedColor = prevButton.HoverBgColor
	prevButton.TextColor = rl.White
	prevButton.BorderColor = rl.NewColor(0, 200, 255, 255)
	prevButton.ShadowColor = rl.Blank
	prevButton.BorderWidth = 2
	prevButton.BorderRadius = 0.25
	prevButton.Render()
	if prevButton.IsClicked && ic.Manager.CurrentPage > 0 && !ic.Manager.IsModalOpen {
		ic.Manager.PrevPage()
	}

	nextButton := NewButton(panelX+panelW-50, panelY+panelH/2-25, 36, 50, ">", 28)
	nextButton.BgColor = rl.NewColor(0, 100, 180, 255)
	nextButton.HoverBgColor = rl.NewColor(0, 150, 220, 255)
	nextButton.ClickedColor = nextButton.HoverBgColor
	nextButton.TextColor = rl.White
	nextButton.BorderColor = rl.NewColor(0, 200, 255, 255)
	nextButton.ShadowColor = rl.Blank
	nextButton.BorderWidth = 2
	nextButton.BorderRadius = 0.25
	nextButton.Render()
	if nextButton.IsClicked && ic.Manager.CurrentPage < ic.Manager.GetTotalPages()-1 && !ic.Manager.IsModalOpen {
		ic.Manager.NextPage()
	}

	if ic.Manager.IsModalOpen && ic.Manager.SelectedItemIndex >= 0 && ic.Manager.SelectedItemIndex < len(ic.Manager.Items) {
		selectedItem := ic.Manager.Items[ic.Manager.SelectedItemIndex]

		rl.DrawRectangle(0, 0, winWidth, winHeight, rl.NewColor(0, 0, 0, 180))

		modalW := float32(320)
		modalH := float32(340)
		modalX := (float32(winWidth) - modalW) / 2
		modalY := (float32(winHeight) - modalH) / 2
		modalRect := rl.NewRectangle(modalX, modalY, modalW, modalH)
		if rl.IsKeyPressed(rl.KeyEscape) {
			ic.Manager.CloseModal()
			return
		}

		rl.DrawRectangleRounded(modalRect, 0.12, 12, rl.NewColor(18, 22, 34, 255))
		rl.DrawRectangleRoundedLinesEx(modalRect, 0.12, 12, 3, rl.Gold)

		modalCloseButton := NewButton(modalX+modalW-44, modalY+8, 34, 34, "X", 14)
		modalCloseButton.BgColor = rl.NewColor(180, 50, 50, 255)
		modalCloseButton.HoverBgColor = rl.Red
		modalCloseButton.ClickedColor = modalCloseButton.HoverBgColor
		modalCloseButton.TextColor = rl.White
		modalCloseButton.BorderColor = rl.White
		modalCloseButton.ShadowColor = rl.Blank
		modalCloseButton.BorderWidth = 1
		modalCloseButton.BorderRadius = 0.3
		modalCloseButton.Render()
		if modalCloseButton.IsClicked {
			ic.Manager.CloseModal()
			return
		}

		imgSize := float32(140)
		imgX := modalX + (modalW-imgSize)/2
		imgY := modalY + 35
		imgRect := rl.NewRectangle(imgX, imgY, imgSize, imgSize)

		rl.DrawRectangleRounded(imgRect, 0.15, 8, rl.NewColor(30, 35, 50, 255))
		rl.DrawRectangleRoundedLinesEx(imgRect, 0.15, 8, 2, rl.NewColor(0, 180, 255, 255))

		tex := ic.GetTexture(selectedItem.Type)
		if tex.ID > 0 {
			srcRec := rl.NewRectangle(0, 0, float32(tex.Width), float32(tex.Height))
			rl.DrawTexturePro(tex, srcRec, imgRect, rl.Vector2{X: 0, Y: 0}, 0, rl.White)
		}

		nameW := float32(rl.MeasureText(selectedItem.Name, 20))
		rl.DrawText(selectedItem.Name, int32(modalX+(modalW-nameW)/2), int32(imgY+imgSize+15), 20, rl.RayWhite)

		countText := fmt.Sprintf("Quantity: %d", selectedItem.Count)
		countW := float32(rl.MeasureText(countText, 14))
		rl.DrawText(countText, int32(modalX+(modalW-countW)/2), int32(imgY+imgSize+42), 14, rl.Gold)

		btnW := float32(110)
		btnH := float32(40)
		btnY := modalY + modalH - 60

		takeButton := NewButton(modalX+30, btnY, btnW, btnH, "TAKE", 16)
		takeButton.BgColor = rl.NewColor(0, 150, 80, 255)
		takeButton.HoverBgColor = rl.NewColor(0, 200, 100, 255)
		takeButton.ClickedColor = takeButton.HoverBgColor
		takeButton.TextColor = rl.White
		takeButton.BorderColor = rl.White
		takeButton.ShadowColor = rl.Blank
		takeButton.BorderWidth = 2
		takeButton.BorderRadius = 0.25
		takeButton.Render()
		if takeButton.IsClicked {
			ic.Manager.TakeItem(ic.Manager.SelectedItemIndex)
		}

		dropButton := NewButton(modalX+modalW-30-btnW, btnY, btnW, btnH, "DROP", 16)
		dropButton.BgColor = rl.NewColor(170, 40, 40, 255)
		dropButton.HoverBgColor = rl.NewColor(220, 50, 50, 255)
		dropButton.ClickedColor = dropButton.HoverBgColor
		dropButton.TextColor = rl.White
		dropButton.BorderColor = rl.White
		dropButton.ShadowColor = rl.Blank
		dropButton.BorderWidth = 2
		dropButton.BorderRadius = 0.25
		dropButton.Render()
		if dropButton.IsClicked {
			ic.Manager.DropItem(ic.Manager.SelectedItemIndex)
		}
	}
}
