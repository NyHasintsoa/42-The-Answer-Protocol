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
	NPCPreviews  map[string]service.INpc
}

func NewRoomDetailsComponent(mgr *service.RoomDetailsManager, pathResolver *utils.PathResolver) *RoomDetailsComponent {
	return &RoomDetailsComponent{
		Manager:      mgr,
		PathResolver: pathResolver,
		ItemComp:     NewItemComponent(pathResolver),
		NPCPreviews:  make(map[string]service.INpc),
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

func (rdc *RoomDetailsComponent) getOrCreateNPCPreview(npc service.RoomNPCDetail) service.INpc {
	if preview, exists := rdc.NPCPreviews[npc.Key]; exists {
		return preview
	}

	var npcComp service.INpc
	var err error
	scale := float32(0.15)

	switch npc.Kind {
	case "guard":
		npcComp, err = NewGuardCharacter(rdc.PathResolver, "npc/guard", 200, 200, scale)
	case "seller":
		npcComp, err = NewSellerCharacter(rdc.PathResolver, "npc/seller", 200, 200, scale)
	default:
		npcComp, err = NewPersonCharacter(rdc.PathResolver, "npc/person", 200, 200, scale)
	}

	if err != nil {
		return nil
	}
	if npcComp != nil {
		rdc.NPCPreviews[npc.Key] = npcComp
	}
	return npcComp
}

func (rdc *RoomDetailsComponent) Render(areaX, areaY, areaW, areaH float32) {
	if rdc.Manager == nil || !rdc.Manager.IsOpen {
		return
	}

	dt := rl.GetFrameTime()
	mousePos := rl.GetMousePosition()
	mouseClicked := rl.IsMouseButtonPressed(rl.MouseLeftButton)

	cardW := float32(900)
	if cardW > areaW-30 {
		cardW = areaW - 30
	}
	cardH := float32(480)
	if cardH > areaH-40 {
		cardH = areaH - 40
	}
	cardX := areaX + (areaW-cardW)/2
	cardY := areaY + (areaH-cardH)/2
	cardRect := rl.NewRectangle(cardX, cardY, cardW, cardH)

	rl.DrawRectangleRounded(cardRect, 0.02, 8, rl.NewColor(245, 247, 250, 255))
	rl.DrawRectangleRoundedLinesEx(cardRect, 0.02, 8, 2, rl.NewColor(180, 190, 200, 255))

	rl.DrawText("ROOM DETAILS", int32(cardX+(cardW-float32(rl.MeasureText("ROOM DETAILS", 16)))/2), int32(cardY+14), 16, rl.NewColor(40, 50, 60, 255))

	colW := (cardW - 35) / 2
	colH := cardH - 52
	colY := cardY + 40

	tealBtnColor := rl.NewColor(24, 160, 178, 255)
	tealBtnHover := rl.NewColor(18, 130, 146, 255)

	leftColX := cardX + 12
	leftColRect := rl.NewRectangle(leftColX, colY, colW, colH)
	rl.DrawRectangleRounded(leftColRect, 0.02, 6, rl.NewColor(238, 242, 246, 255))
	rl.DrawRectangleRoundedLinesEx(leftColRect, 0.02, 6, 1, rl.NewColor(200, 210, 220, 255))

	rl.DrawText("Items in Room", int32(leftColX+(colW-float32(rl.MeasureText("Items in Room", 14)))/2), int32(colY+12), 14, rl.NewColor(50, 60, 75, 255))
	rl.DrawLineEx(rl.NewVector2(leftColX, colY+34), rl.NewVector2(leftColX+colW, colY+34), 1, rl.NewColor(210, 220, 230, 255))

	if len(rdc.Manager.Items) == 0 {
		emptyMsg := "No items in this room."
		rl.DrawText(emptyMsg, int32(leftColX+(colW-float32(rl.MeasureText(emptyMsg, 13)))/2), int32(colY+colH/2-10), 13, rl.Gray)
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

			rl.DrawText(item.Name, int32(leftColX+76), int32(itemY+2), 14, rl.NewColor(20, 25, 30, 255))
			rl.DrawText(item.Description, int32(leftColX+76), int32(itemY+20), 10, rl.NewColor(100, 110, 120, 255))

			btnW := float32(90)
			btnH := float32(26)
			btnX := leftColX + (colW-btnW)/2
			btnY := itemY + 40
			btnRect := rl.NewRectangle(btnX, btnY, btnW, btnH)

			btnColor := tealBtnColor
			if rl.CheckCollisionPointRec(mousePos, btnRect) {
				btnColor = tealBtnHover
				if mouseClicked {
					fmt.Printf("[ACTION LOG] TAKE clicked -> Item: %s (ID: %s)\n", item.Name, item.ID)
					rdc.Manager.TakeItem(item.ID)
				}
			}

			rl.DrawRectangleRounded(btnRect, 0.2, 4, btnColor)
			rl.DrawText("TAKE", int32(btnX+(btnW-float32(rl.MeasureText("TAKE", 12)))/2), int32(btnY+7), 12, rl.White)

			itemY += 95
		}
	}

	rightColX := cardX + colW + 22
	rightColRect := rl.NewRectangle(rightColX, colY, colW, colH)
	rl.DrawRectangleRounded(rightColRect, 0.02, 6, rl.NewColor(238, 242, 246, 255))
	rl.DrawRectangleRoundedLinesEx(rightColRect, 0.02, 6, 1, rl.NewColor(200, 210, 220, 255))

	rl.DrawText("NPC in Room", int32(rightColX+(colW-float32(rl.MeasureText("NPC in Room", 14)))/2), int32(colY+12), 14, rl.NewColor(50, 60, 75, 255))
	rl.DrawLineEx(rl.NewVector2(rightColX, colY+34), rl.NewVector2(rightColX+colW, colY+34), 1, rl.NewColor(210, 220, 230, 255))

	if len(rdc.Manager.NPCs) == 0 {
		emptyMsg := "No NPCs in this room."
		rl.DrawText(emptyMsg, int32(rightColX+(colW-float32(rl.MeasureText(emptyMsg, 13)))/2), int32(colY+colH/2-10), 13, rl.Gray)
	} else {
		npcY := colY + 45
		for _, npc := range rdc.Manager.NPCs {
			npcComp := rdc.getOrCreateNPCPreview(npc)
			previewX := rightColX + 12
			previewY := npcY + 42
			infoX := rightColX + 86
			if npcComp != nil {
				npcComp.Update(dt)
				if npcChar, ok := npcComp.(*NpcCharacter); ok {
					npcChar.Position = rl.NewVector2(previewX, previewY)
					npcChar.ShowLabel = false
				}
				npcComp.Render()
			}

			rl.DrawText(npc.Name, int32(infoX), int32(npcY+42), 14, rl.NewColor(20, 25, 30, 255))
			rl.DrawText(npc.Description, int32(infoX), int32(npcY+62), 10, rl.NewColor(100, 110, 120, 255))

			btnW := float32(80)
			btnH := float32(26)
			btnY := npcY + 82

			if npc.HasQuest {
				buttonGroupW := btnW*2 + 10
				talkX := infoX + (colW-(infoX-rightColX)-buttonGroupW)/2
				questX := talkX + btnW + 10

				talkRect := rl.NewRectangle(talkX, btnY, btnW, btnH)
				talkBg := tealBtnColor
				if rl.CheckCollisionPointRec(mousePos, talkRect) {
					talkBg = tealBtnHover
					if mouseClicked {
						fmt.Printf("[ACTION LOG] TALK button clicked -> NPC: %s (%s)\n", npc.Name, npc.Key)
						rdc.Manager.TalkToNPC(npc.Key, npc.Name)
					}
				}
				rl.DrawRectangleRounded(talkRect, 0.2, 4, talkBg)
				rl.DrawText("TALK", int32(talkX+(btnW-float32(rl.MeasureText("TALK", 12)))/2), int32(btnY+7), 12, rl.White)

				questRect := rl.NewRectangle(questX, btnY, btnW, btnH)
				questBg := tealBtnColor
				if rl.CheckCollisionPointRec(mousePos, questRect) {
					questBg = tealBtnHover
					if mouseClicked {
						fmt.Printf("[ACTION LOG] QUEST button clicked -> NPC: %s (Quest ID: %s)\n", npc.Name, npc.QuestID)
						rdc.Manager.OpenQuest(npc.Key, npc.QuestID)
					}
				}
				rl.DrawRectangleRounded(questRect, 0.2, 4, questBg)
				rl.DrawText("QUEST", int32(questX+(btnW-float32(rl.MeasureText("QUEST", 12)))/2), int32(btnY+7), 12, rl.White)
			} else {
				talkX := infoX + (colW-(infoX-rightColX)-btnW)/2
				talkRect := rl.NewRectangle(talkX, btnY, btnW, btnH)
				talkBg := tealBtnColor
				if rl.CheckCollisionPointRec(mousePos, talkRect) {
					talkBg = tealBtnHover
					if mouseClicked {
						fmt.Printf("[ACTION LOG] TALK button clicked -> NPC: %s (%s)\n", npc.Name, npc.Key)
						rdc.Manager.TalkToNPC(npc.Key, npc.Name)
					}
				}
				rl.DrawRectangleRounded(talkRect, 0.2, 4, talkBg)
				rl.DrawText("TALK", int32(talkX+(btnW-float32(rl.MeasureText("TALK", 12)))/2), int32(btnY+7), 12, rl.White)
			}

			npcY += 120
		}
	}
}
