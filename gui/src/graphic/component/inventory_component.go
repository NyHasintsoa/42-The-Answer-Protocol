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

	tex := rl.LoadTexture(imagePath)
	ic.textures[itemType] = tex
	return tex
}

func (ic *InventoryComponent) Unload() {
	for _, tex := range ic.textures {
		rl.UnloadTexture(tex)
	}
	ic.textures = make(map[enums.InventoryItemType]rl.Texture2D)
}

func (ic *InventoryComponent) Render(winWidth, winHeight int32) {
	if ic.Manager == nil {
		return
	}

	mousePos := rl.GetMousePosition()
	mouseClicked := rl.IsMouseButtonPressed(rl.MouseLeftButton)

	
	btnX := float32(20)
	btnY := float32(winHeight - 130)
	btnW := float32(110)
	btnH := float32(40)
	btnRect := rl.NewRectangle(btnX, btnY, btnW, btnH)

	btnBg := rl.NewColor(30, 35, 55, 240)
	btnBorder := rl.NewColor(0, 180, 255, 255)
	if rl.CheckCollisionPointRec(mousePos, btnRect) {
		btnBg = rl.NewColor(50, 70, 110, 255)
		btnBorder = rl.Gold
	}

	rl.DrawRectangleRounded(btnRect, 0.25, 8, btnBg)
	rl.DrawRectangleRoundedLinesEx(btnRect, 0.25, 8, 2, btnBorder)
	rl.DrawText("BAG [I]", int32(btnX+22), int32(btnY+11), 18, rl.RayWhite)

	if !ic.Manager.IsOpen {
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

	
	closeBtn := rl.NewRectangle(panelX+panelW-42, panelY+12, 30, 30)
	closeCol := rl.NewColor(180, 50, 50, 255)
	if rl.CheckCollisionPointRec(mousePos, closeBtn) {
		closeCol = rl.Red
	}
	rl.DrawRectangleRounded(closeBtn, 0.3, 8, closeCol)
	rl.DrawRectangleRoundedLinesEx(closeBtn, 0.3, 8, 1, rl.White)
	rl.DrawText("X", int32(panelX+panelW-33), int32(panelY+17), 18, rl.White)

	
	slotSize := float32(115)
	spacing := float32(18)
	gridStartX := panelX + (panelW-3*slotSize-2*spacing)/2
	gridStartY := panelY + 80

	startIndex := ic.Manager.CurrentPage * ic.Manager.ItemsPerPage
	endIndex := startIndex + ic.Manager.ItemsPerPage
	if endIndex > len(ic.Manager.Items) {
		endIndex = len(ic.Manager.Items)
	}

	for i := 0; i < 9; i++ {
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

	
	prevBtn := rl.NewRectangle(panelX+14, panelY+panelH/2-25, 36, 50)
	prevCol := rl.NewColor(30, 40, 60, 255)
	if ic.Manager.CurrentPage > 0 {
		if rl.CheckCollisionPointRec(mousePos, prevBtn) {
			prevCol = rl.NewColor(0, 150, 220, 255)
		} else {
			prevCol = rl.NewColor(0, 100, 180, 255)
		}
	}
	rl.DrawRectangleRounded(prevBtn, 0.25, 8, prevCol)
	rl.DrawRectangleRoundedLinesEx(prevBtn, 0.25, 8, 2, rl.NewColor(0, 200, 255, 255))
	p1 := rl.Vector2{X: panelX + 24, Y: panelY + panelH/2}
	p2 := rl.Vector2{X: panelX + 40, Y: panelY + panelH/2 - 14}
	p3 := rl.Vector2{X: panelX + 40, Y: panelY + panelH/2 + 14}
	rl.DrawTriangle(p1, p3, p2, rl.White)

	
	nextBtn := rl.NewRectangle(panelX+panelW-50, panelY+panelH/2-25, 36, 50)
	nextCol := rl.NewColor(30, 40, 60, 255)
	if ic.Manager.CurrentPage < ic.Manager.GetTotalPages()-1 {
		if rl.CheckCollisionPointRec(mousePos, nextBtn) {
			nextCol = rl.NewColor(0, 150, 220, 255)
		} else {
			nextCol = rl.NewColor(0, 100, 180, 255)
		}
	}
	rl.DrawRectangleRounded(nextBtn, 0.25, 8, nextCol)
	rl.DrawRectangleRoundedLinesEx(nextBtn, 0.25, 8, 2, rl.NewColor(0, 200, 255, 255))
	np1 := rl.Vector2{X: panelX + panelW - 24, Y: panelY + panelH/2}
	np2 := rl.Vector2{X: panelX + panelW - 40, Y: panelY + panelH/2 - 14}
	np3 := rl.Vector2{X: panelX + panelW - 40, Y: panelY + panelH/2 + 14}
	rl.DrawTriangle(np1, np2, np3, rl.White)

	
	if ic.Manager.IsModalOpen && ic.Manager.SelectedItemIndex >= 0 && ic.Manager.SelectedItemIndex < len(ic.Manager.Items) {
		selectedItem := ic.Manager.Items[ic.Manager.SelectedItemIndex]

		
		rl.DrawRectangle(0, 0, winWidth, winHeight, rl.NewColor(0, 0, 0, 180))

		modalW := float32(320)
		modalH := float32(340)
		modalX := (float32(winWidth) - modalW) / 2
		modalY := (float32(winHeight) - modalH) / 2
		modalRect := rl.NewRectangle(modalX, modalY, modalW, modalH)

		rl.DrawRectangleRounded(modalRect, 0.12, 12, rl.NewColor(18, 22, 34, 255))
		rl.DrawRectangleRoundedLinesEx(modalRect, 0.12, 12, 3, rl.Gold)

		
		mCloseBtn := rl.NewRectangle(modalX+modalW-36, modalY+10, 26, 26)
		if mouseClicked && rl.CheckCollisionPointRec(mousePos, mCloseBtn) {
			ic.Manager.CloseModal()
			return
		}
		rl.DrawRectangleRounded(mCloseBtn, 0.3, 6, rl.NewColor(180, 50, 50, 255))
		rl.DrawRectangleRoundedLinesEx(mCloseBtn, 0.3, 6, 1, rl.White)
		rl.DrawText("X", int32(modalX+modalW-28), int32(modalY+14), 14, rl.White)

		
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

		takeBtn := rl.NewRectangle(modalX+30, btnY, btnW, btnH)
		takeBg := rl.NewColor(0, 150, 80, 255)
		if rl.CheckCollisionPointRec(mousePos, takeBtn) {
			takeBg = rl.NewColor(0, 200, 100, 255)
			if mouseClicked {
				ic.Manager.TakeItem(ic.Manager.SelectedItemIndex)
			}
		}
		rl.DrawRectangleRounded(takeBtn, 0.25, 8, takeBg)
		rl.DrawRectangleRoundedLinesEx(takeBtn, 0.25, 8, 2, rl.White)
		tW := float32(rl.MeasureText("TAKE", 16))
		rl.DrawText("TAKE", int32(takeBtn.X+(btnW-tW)/2), int32(btnY+12), 16, rl.White)

		dropBtn := rl.NewRectangle(modalX+modalW-30-btnW, btnY, btnW, btnH)
		dropBg := rl.NewColor(170, 40, 40, 255)
		if rl.CheckCollisionPointRec(mousePos, dropBtn) {
			dropBg = rl.NewColor(220, 50, 50, 255)
			if mouseClicked {
				ic.Manager.DropItem(ic.Manager.SelectedItemIndex)
			}
		}
		rl.DrawRectangleRounded(dropBtn, 0.25, 8, dropBg)
		rl.DrawRectangleRoundedLinesEx(dropBtn, 0.25, 8, 2, rl.White)
		dW := float32(rl.MeasureText("DROP", 16))
		rl.DrawText("DROP", int32(dropBtn.X+(btnW-dW)/2), int32(btnY+12), 16, rl.White)
	}
}