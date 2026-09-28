package component

import (
	"fmt"
	"strings"

	"tap-gui/src/service"
	"tap-gui/src/utils"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type RoomDetailsComponent struct {
	Manager      *service.RoomDetailsManager
	PathResolver *utils.PathResolver
	ItemComp     *ItemComponent
	NPCPreviews  map[string]INpc
}

func NewRoomDetailsComponent(mgr *service.RoomDetailsManager, pathResolver *utils.PathResolver) *RoomDetailsComponent {
	return &RoomDetailsComponent{
		Manager:      mgr,
		PathResolver: pathResolver,
		ItemComp:     NewItemComponent(pathResolver),
		NPCPreviews:  make(map[string]INpc),
	}
}

func (rdc *RoomDetailsComponent) Unload() {
	if rdc.ItemComp != nil {
		rdc.ItemComp.Unload()
	}
	for _, npc := range rdc.NPCPreviews {
		if npc != nil {
			npc.UnloadAnimations()
		}
	}
}

func (rdc *RoomDetailsComponent) getOrCreateNPCPreview(npc service.NPCConfig, x float32, y float32) INpc {
	if preview, exists := rdc.NPCPreviews[npc.Key]; exists {
		return preview
	}

	var npcComp INpc
	var err error
	scale := float32(0.15)

	imageType := strings.ToLower(npc.ImageType)
	if imageType == "" {
		imageType = strings.ToLower(npc.Type)
	}

	switch imageType {
	case "boss":
		npcComp, err = NewEnemyCharacter(rdc.PathResolver, "characters/boss", x, y, scale)
	case "gobelin", "goblin":
		npcComp, err = NewEnemyCharacter(rdc.PathResolver, "characters/gobelin", x, y, scale)
	case "guard":
		npcComp, err = NewGuardCharacter(rdc.PathResolver, "npc/guard", x, y, scale)
	case "merchant", "seller":
		npcComp, err = NewSellerCharacter(rdc.PathResolver, "npc/merchant", x, y, scale)
	default:
		relDir := "npc/" + imageType
		npcComp, err = NewPersonCharacter(rdc.PathResolver, relDir, x, y, scale)
		if err != nil {
			npcComp, err = NewPersonCharacter(rdc.PathResolver, "npc/person", x, y, scale)
		}
	}

	if err != nil {
		return nil
	}
	if npcComp != nil {
		rdc.NPCPreviews[npc.Key] = npcComp
	}
	return npcComp
}

func (rdc *RoomDetailsComponent) Render(rect rl.Rectangle) {
	if rdc.Manager == nil || !rdc.Manager.IsOpen {
		return
	}

	dt := rl.GetFrameTime()

	rl.DrawRectangleRounded(rect, 0.02, 6, utils.ThemeSurface)
	rl.DrawRectangleRoundedLinesEx(rect, 0.02, 6, 2, utils.ThemeBorder)

	rl.DrawText("ROOM DETAILS", int32(rect.X+(rect.Width-float32(rl.MeasureText("ROOM DETAILS", 16)))/2), int32(rect.Y+14), 16, utils.ThemeText)

	colW := (rect.Width - 35) / 2
	colH := rect.Height - 52
	colY := rect.Y + 40

	leftColX := rect.X + 12
	leftColRect := rl.NewRectangle(leftColX, colY, colW, colH)
	rl.DrawRectangleRounded(leftColRect, 0.02, 6, utils.ThemeSurfaceInset)
	rl.DrawRectangleRoundedLinesEx(leftColRect, 0.02, 6, 1, utils.ThemeBorder)

	rl.DrawText("Items in Room", int32(leftColX+(colW-float32(rl.MeasureText("Items in Room", 14)))/2), int32(colY+12), 14, utils.ThemeText)
	rl.DrawLineEx(rl.NewVector2(leftColX, colY+34), rl.NewVector2(leftColX+colW, colY+34), 1, utils.ThemeBorder)

	if len(rdc.Manager.Items) == 0 {
		emptyMsg := "No items in this room."
		rl.DrawText(emptyMsg, int32(leftColX+(colW-float32(rl.MeasureText(emptyMsg, 13)))/2), int32(colY+colH/2-10), 13, utils.ThemeTextMuted)
	} else {
		itemY := colY + 45
		for _, item := range rdc.Manager.Items {
			parts := strings.Split(item.ID, ":")
			iconName := item.ID
			if len(parts) > 1 {
				iconName = parts[1]
			}

			iconRect := rl.NewRectangle(leftColX+18, itemY+5, 48, 48)
			rl.DrawRectangleRounded(iconRect, 0.2, 4, rl.NewColor(30, 32, 40, 255))
			rl.DrawRectangleRoundedLinesEx(iconRect, 0.2, 4, 1, rl.NewColor(80, 90, 105, 255))

			if rdc.ItemComp != nil {
				rdc.ItemComp.RenderItemIcon(iconName, iconRect)
			}

			rl.DrawText(item.Name, int32(leftColX+76), int32(itemY+2), 14, utils.ThemeText)
			rl.DrawText(item.Description, int32(leftColX+76), int32(itemY+20), 10, utils.ThemeTextMuted)

			btnW := float32(90)
			btnH := float32(26)
			btnX := leftColX + (colW-btnW)/2
			btnY := itemY + 40

			takeButton := NewButton(btnX, btnY, btnW, btnH, "TAKE", 12)
			takeButton.BgColor = rl.NewColor(24, 160, 178, 255)
			takeButton.HoverBgColor = rl.NewColor(18, 130, 146, 255)
			takeButton.ClickedColor = takeButton.HoverBgColor
			takeButton.TextColor = rl.White
			takeButton.BorderColor = rl.Blank
			takeButton.ShadowColor = rl.Blank
			takeButton.BorderWidth = 0
			takeButton.BorderRadius = 0.2
			takeButton.Render()

			if takeButton.IsClicked {
				fmt.Printf("[ACTION LOG] TAKE clicked -> Item: %s (ID: %s)\n", item.Name, item.ID)
				rdc.Manager.TakeItem(item.ID)
			}

			itemY += 95
		}
	}

	rightColX := rect.X + colW + 22
	rightColRect := rl.NewRectangle(rightColX, colY, colW, colH)
	rl.DrawRectangleRounded(rightColRect, 0.02, 6, utils.ThemeSurfaceInset)
	rl.DrawRectangleRoundedLinesEx(rightColRect, 0.02, 6, 1, utils.ThemeBorder)

	rl.DrawText("NPC in Room", int32(rightColX+(colW-float32(rl.MeasureText("NPC in Room", 14)))/2), int32(colY+12), 14, utils.ThemeText)
	rl.DrawLineEx(rl.NewVector2(rightColX, colY+34), rl.NewVector2(rightColX+colW, colY+34), 1, utils.ThemeBorder)

	if len(rdc.Manager.NPCs) == 0 {
		emptyMsg := "No NPCs in this room."
		rl.DrawText(emptyMsg, int32(rightColX+(colW-float32(rl.MeasureText(emptyMsg, 13)))/2), int32(colY+colH/2-10), 13, utils.ThemeTextMuted)
	} else {
		npcY := colY + 45
		for _, npc := range rdc.Manager.NPCs {
			npcComp := rdc.getOrCreateNPCPreview(npc, rightColX, npcY)
			previewX := rightColX + 12
			previewY := npcY + 42
			infoX := rightColX + 100

			if npcComp != nil {
				npcComp.Update(dt)
				npcComp.SetPosition(previewX, previewY)
				npcComp.Render()
			}

			rl.DrawText(npc.Name, int32(infoX), int32(npcY+42), 14, utils.ThemeText)
			rl.DrawText(npc.Description, int32(infoX), int32(npcY+62), 10, utils.ThemeTextMuted)

			btnW := float32(90)
			btnH := float32(26)
			btnX := infoX
			btnY := npcY + 82

			actionsButton := NewButton(btnX, btnY, btnW, btnH, "ACTIONS", 12)
			actionsButton.BgColor = rl.NewColor(24, 160, 178, 255)
			actionsButton.HoverBgColor = rl.NewColor(18, 130, 146, 255)
			actionsButton.ClickedColor = actionsButton.HoverBgColor
			actionsButton.TextColor = rl.White
			actionsButton.BorderColor = rl.Blank
			actionsButton.ShadowColor = rl.Blank
			actionsButton.BorderWidth = 0
			actionsButton.BorderRadius = 0.2
			actionsButton.Render()
			if actionsButton.IsClicked {
				rdc.Manager.OpenActions(npc)
			}

			npcY += 120
		}
	}
}
