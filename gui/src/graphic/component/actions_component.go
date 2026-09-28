package component

import (
	"fmt"
	"strings"

	"tap-gui/src/service"
	"tap-gui/src/utils"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type ActionComponent struct {
	Manager      *service.ActionManager
	PathResolver *utils.PathResolver
	NPCPreviews  map[string]INpc
	LawnTexture  rl.Texture2D
}

func NewActionComponent(mgr *service.ActionManager, pathResolver *utils.PathResolver) *ActionComponent {
	component := &ActionComponent{
		Manager:      mgr,
		PathResolver: pathResolver,
		NPCPreviews:  make(map[string]INpc),
	}
	if pathResolver != nil && pathResolver.ResourceManager != nil {
		component.LawnTexture = pathResolver.ResourceManager.Load(pathResolver.Resolve("ground", "lawn.png"))
	}
	return component
}

func (dc *ActionComponent) Unload() {
	for _, npc := range dc.NPCPreviews {
		if npc != nil {
			npc.UnloadAnimations()
		}
	}
}

func (dc *ActionComponent) getOrCreateNPCPreview(npc service.NPCConfig, x float32, y float32) INpc {
	if preview, exists := dc.NPCPreviews[npc.Key]; exists {
		return preview
	}

	var npcComp INpc
	var err error
	scale := float32(0.25)

	imageType := strings.ToLower(npc.ImageType)
	if imageType == "" {
		imageType = strings.ToLower(npc.Type)
	}

	switch imageType {
	case "boss":
		npcComp, err = NewEnemyCharacter(dc.PathResolver, "characters/boss", x, y, scale)
	case "gobelin", "goblin":
		npcComp, err = NewEnemyCharacter(dc.PathResolver, "characters/gobelin", x, y, scale)
	case "guard":
		npcComp, err = NewGuardCharacter(dc.PathResolver, "npc/guard", x, y, scale)
	case "merchant", "seller":
		npcComp, err = NewSellerCharacter(dc.PathResolver, "npc/merchant", x, y, scale)
	default:
		relDir := "npc/" + imageType
		npcComp, err = NewPersonCharacter(dc.PathResolver, relDir, x, y, scale)
		if err != nil {
			npcComp, err = NewPersonCharacter(dc.PathResolver, "npc/person", x, y, scale)
		}
	}

	if err != nil {
		return nil
	}
	if npcComp != nil {
		dc.NPCPreviews[npc.Key] = npcComp
	}
	return npcComp
}

func (dc *ActionComponent) Render(winWidth, winHeight int32) {
	if dc.Manager == nil || !dc.Manager.IsOpen {
		return
	}

	dt := rl.GetFrameTime()

	modalWidth := float32(900)
	modalHeight := float32(600)
	modalX := (float32(winWidth) - modalWidth) / 2
	modalY := (float32(winHeight) - modalHeight) / 2

	padding := modalWidth * 0.018
	headerH := float32(42)
	respHeaderH := float32(26)

	totalAvailableH := modalHeight - headerH - respHeaderH - (padding * 3.5)
	contentH := totalAvailableH * 0.56
	responseAreaH := totalAvailableH * 0.44

	rl.DrawRectangle(0, 0, winWidth, winHeight, rl.NewColor(0, 0, 0, 180))

	cyanBorder := rl.NewColor(0, 184, 212, 255)
	darkBg := rl.NewColor(10, 17, 28, 245)
	boxBg := rl.NewColor(8, 14, 23, 255)

	modalRect := rl.NewRectangle(modalX, modalY, modalWidth, modalHeight)
	rl.DrawRectangleRounded(modalRect, 0.03, 8, darkBg)
	rl.DrawRectangleRoundedLinesEx(modalRect, 0.03, 8, 2, cyanBorder)

	titleText := fmt.Sprintf("ACTIONS: %s", strings.ToUpper(dc.Manager.NPC.Name))
	rl.DrawText(titleText, int32(modalX+padding), int32(modalY+12), 18, cyanBorder)

	closeBtnW := float32(36)
	closeBtnH := float32(24)
	closeBtnX := modalX + modalWidth - closeBtnW - padding
	closeBtnY := modalY + 10

	closeBtn := NewButton(closeBtnX, closeBtnY, closeBtnW, closeBtnH, " X ", 12)
	closeBtn.BgColor = rl.NewColor(160, 20, 20, 255)
	closeBtn.HoverBgColor = rl.NewColor(200, 30, 30, 255)
	closeBtn.TextColor = rl.White
	closeBtn.BorderColor = rl.White
	closeBtn.BorderWidth = 1
	closeBtn.Render()

	if closeBtn.IsClicked || rl.IsKeyPressed(rl.KeyEscape) {
		dc.Manager.Close()
	}

	contentX := modalX + padding
	contentY := modalY + headerH + 6
	contentW := modalWidth - (padding * 2)

	contentRect := rl.NewRectangle(contentX, contentY, contentW, contentH)
	rl.DrawRectangleRounded(contentRect, 0.02, 6, darkBg)
	rl.DrawRectangleRoundedLinesEx(contentRect, 0.02, 6, 1, cyanBorder)

	portraitW := contentW * 0.22
	if portraitW < 150 {
		portraitW = 150
	}
	portraitH := contentH - (padding * 1.6)
	portraitX := contentX + padding
	portraitY := contentY + (padding * 0.8)

	portraitRect := rl.NewRectangle(portraitX, portraitY, portraitW, portraitH)
	if dc.LawnTexture.ID > 0 {
		srcRect := rl.NewRectangle(0, 0, float32(dc.LawnTexture.Width), float32(dc.LawnTexture.Height))
		rl.DrawTexturePro(dc.LawnTexture, srcRect, portraitRect, rl.Vector2Zero(), 0, rl.White)
	} else {
		rl.DrawRectangleRounded(portraitRect, 0.03, 6, boxBg)
	}
	rl.DrawRectangleRoundedLinesEx(portraitRect, 0.03, 6, 1, cyanBorder)

	previewChar := dc.getOrCreateNPCPreview(dc.Manager.NPC, portraitX+portraitW/2, portraitY+portraitH-30)
	if previewChar != nil {
		previewChar.SetPosition(portraitX+portraitW/2, portraitY+portraitH-30)
		previewChar.SetAction("Communication")
		previewChar.Update(dt)
		previewChar.Render()
	}

	npcTag := fmt.Sprintf("[ %s ]", dc.Manager.NPC.Name)
	rl.DrawText(npcTag, int32(portraitX+(portraitW-float32(rl.MeasureText(npcTag, 11)))/2), int32(portraitY+portraitH-22), 11, utils.ThemeTextMuted)

	actionX := portraitX + portraitW + padding
	actionY := portraitY
	actionW := contentW - portraitW - (padding * 3)
	actionH := portraitH

	actionRect := rl.NewRectangle(actionX, actionY, actionW, actionH)
	rl.DrawRectangleRounded(actionRect, 0.03, 6, boxBg)
	rl.DrawRectangleRoundedLinesEx(actionRect, 0.03, 6, 1, cyanBorder)

	rl.DrawText(strings.ToUpper(dc.Manager.NPC.Name), int32(actionX+14), int32(actionY+14), 14, cyanBorder)
	rl.DrawLineEx(rl.NewVector2(actionX+14, actionY+32), rl.NewVector2(actionX+actionW-14, actionY+32), 1, rl.NewColor(0, 100, 120, 255))

	drawWrappedText(dc.Manager.CurrentText, int32(actionX+14), int32(actionY+45), int32(actionW-28), 13, utils.ThemeText)

	respHeaderY := contentY + contentH + (padding * 0.8)
	rl.DrawText("YOUR RESPONSES / ACTIONS:", int32(modalX+padding), int32(respHeaderY), 14, cyanBorder)

	respStartY := respHeaderY + respHeaderH
	respW := contentW
	numResponses := float32(len(dc.Manager.Responses))
	if numResponses == 0 {
		numResponses = 1
	}

	spacing := float32(8)
	respH := (responseAreaH - (spacing * (numResponses - 1))) / numResponses
	if respH < 32 {
		respH = 32
	}

	for i, resp := range dc.Manager.Responses {
		itemY := respStartY + float32(i)*(respH+spacing)
		itemRect := rl.NewRectangle(modalX+padding, itemY, respW, respH)

		mousePos := rl.GetMousePosition()
		isHover := rl.CheckCollisionPointRec(mousePos, itemRect)

		itemBg := darkBg
		if isHover {
			itemBg = rl.NewColor(15, 35, 55, 255)
		}

		rl.DrawRectangleRounded(itemRect, 0.08, 6, itemBg)
		rl.DrawRectangleRoundedLinesEx(itemRect, 0.08, 6, 1, cyanBorder)

		numBoxW := respH - 10
		if numBoxW < 24 {
			numBoxW = 24
		}
		numBoxH := respH - 10
		if numBoxH < 24 {
			numBoxH = 24
		}
		numBoxRect := rl.NewRectangle(modalX+padding+6, itemY+5, numBoxW, numBoxH)

		rl.DrawRectangleRounded(numBoxRect, 0.2, 4, rl.NewColor(0, 184, 212, 40))
		rl.DrawRectangleRoundedLinesEx(numBoxRect, 0.2, 4, 1, cyanBorder)

		numStr := fmt.Sprintf("%d", resp.Index)
		rl.DrawText(numStr, int32(numBoxRect.X+(numBoxW-float32(rl.MeasureText(numStr, 14)))/2), int32(numBoxRect.Y+(numBoxH-14)/2), 14, cyanBorder)

		rl.DrawText(resp.Text, int32(modalX+padding+numBoxW+16), int32(itemY+(respH-13)/2), 13, utils.ThemeText)

		if resp.ActionTag != "" {
			tagW := float32(rl.MeasureText(resp.ActionTag, 12))
			tagX := modalX + padding + respW - tagW - 12
			tagColor := rl.NewColor(212, 160, 23, 255)
			if strings.Contains(resp.ActionTag, "Close") {
				tagColor = utils.ThemeTextMuted
			}
			rl.DrawText(resp.ActionTag, int32(tagX), int32(itemY+(respH-12)/2), 12, tagColor)
		}

		if (isHover && rl.IsMouseButtonPressed(rl.MouseButtonLeft)) || rl.IsKeyPressed(int32(rl.KeyOne)+int32(i)) {
			if resp.OnSelect != nil {
				resp.OnSelect()
			} else {
				fmt.Printf("[DIALOGUE] Selected option %d: %s\n", resp.Index, resp.Text)
			}
		}
	}
}

func drawWrappedText(text string, x, y, maxWidth, fontSize int32, color rl.Color) {
	words := strings.Fields(text)
	var currentLine string
	lineY := y

	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if rl.MeasureText(testLine, fontSize) > maxWidth {
			rl.DrawText(currentLine, x, lineY, fontSize, color)
			currentLine = word
			lineY += fontSize + 6
		} else {
			currentLine = testLine
		}
	}

	if currentLine != "" {
		rl.DrawText(currentLine, x, lineY, fontSize, color)
	}
}
