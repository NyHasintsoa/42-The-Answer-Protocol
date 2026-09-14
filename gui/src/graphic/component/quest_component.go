package component

import (
	"fmt"
	"tap-gui/src/service"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type QuestComponent struct {
	Manager *service.QuestManager
	ScrollY float32
}

func NewQuestComponent(mgr *service.QuestManager) *QuestComponent {
	return &QuestComponent{
		Manager: mgr,
		ScrollY: 0,
	}
}

func (qc *QuestComponent) Render(winWidth, winHeight int32) {
	if qc.Manager == nil || !qc.Manager.IsOpen {
		return
	}

	mousePos := rl.GetMousePosition()
	mouseClicked := rl.IsMouseButtonPressed(rl.MouseLeftButton)

	rl.DrawRectangle(0, 0, winWidth, winHeight, rl.NewColor(0, 0, 0, 160))

	modalW := float32(650)
	modalH := float32(520)
	modalX := (float32(winWidth) - modalW) / 2
	modalY := (float32(winHeight) - modalH) / 2
	modalRect := rl.NewRectangle(modalX, modalY, modalW, modalH)

	rl.DrawRectangleRounded(modalRect, 0.05, 12, rl.NewColor(18, 22, 34, 250))
	rl.DrawRectangleRoundedLinesEx(modalRect, 0.05, 12, 2, rl.NewColor(0, 180, 255, 255))

	headerH := float32(55)
	headerRect := rl.NewRectangle(modalX, modalY, modalW, headerH)
	rl.DrawRectangleRounded(headerRect, 0.08, 12, rl.NewColor(14, 18, 28, 255))
	rl.DrawLineEx(rl.NewVector2(modalX, modalY+headerH), rl.NewVector2(modalX+modalW, modalY+headerH), 2, rl.NewColor(0, 180, 255, 255))

	rl.DrawText("NPC QUEST LOG", int32(modalX+24), int32(modalY+16), 20, rl.NewColor(0, 200, 255, 255))

	closeBtn := rl.NewRectangle(modalX+modalW-40, modalY+12, 30, 30)
	closeCol := rl.NewColor(180, 50, 50, 255)
	if rl.CheckCollisionPointRec(mousePos, closeBtn) {
		closeCol = rl.Red
		if mouseClicked {
			qc.Manager.IsOpen = false
		}
	}
	rl.DrawRectangleRounded(closeBtn, 0.3, 8, closeCol)
	rl.DrawRectangleRoundedLinesEx(closeBtn, 0.3, 8, 1, rl.White)
	rl.DrawText("X", int32(modalX+modalW-31), int32(modalY+17), 18, rl.White)

	quests := qc.Manager.Quests
	listAreaY := modalY + headerH + 15
	listAreaH := modalH - headerH - 30
	listAreaRect := rl.NewRectangle(modalX, listAreaY, modalW, listAreaH)

	cardH := float32(75)
	gap := float32(10)
	totalContentH := float32(len(quests))*(cardH+gap) + 10

	maxScroll := totalContentH - listAreaH
	if maxScroll < 0 {
		maxScroll = 0
	}

	if rl.CheckCollisionPointRec(mousePos, listAreaRect) {
		wheel := rl.GetMouseWheelMove()
		if wheel != 0 {
			qc.ScrollY -= wheel * 35
		}
	}

	if qc.ScrollY < 0 {
		qc.ScrollY = 0
	}
	if qc.ScrollY > maxScroll {
		qc.ScrollY = maxScroll
	}

	rl.BeginScissorMode(int32(modalX), int32(listAreaY), int32(modalW), int32(listAreaH))

	currY := listAreaY + 10 - qc.ScrollY
	cardW := modalW - 40
	cardX := modalX + 20

	for _, quest := range quests {
		cardRect := rl.NewRectangle(cardX, currY, cardW, cardH)

		rl.DrawRectangleRounded(cardRect, 0.15, 8, rl.NewColor(26, 32, 48, 255))
		rl.DrawRectangleRoundedLinesEx(cardRect, 0.15, 8, 1, rl.NewColor(45, 60, 85, 255))

		rl.DrawText(quest.Title, int32(cardX+16), int32(currY+12), 16, rl.White)
		npcText := fmt.Sprintf("NPC: %s", quest.NPCName)
		rl.DrawText(npcText, int32(cardX+16), int32(currY+32), 12, rl.NewColor(120, 160, 200, 255))

		barX := cardX + 16
		barY := currY + 50
		barW := cardW - 160
		barH := float32(14)

		progressRatio := float32(0)
		if quest.TargetProgress > 0 {
			progressRatio = float32(quest.CurrentProgress) / float32(quest.TargetProgress)
		}
		if progressRatio > 1.0 {
			progressRatio = 1.0
		}

		barBgRect := rl.NewRectangle(barX, barY, barW, barH)
		barFillRect := rl.NewRectangle(barX, barY, barW*progressRatio, barH)

		rl.DrawRectangleRounded(barBgRect, 0.4, 6, rl.NewColor(15, 20, 30, 255))
		if progressRatio > 0 {
			rl.DrawRectangleRounded(barFillRect, 0.4, 6, rl.NewColor(0, 180, 255, 255))
		}
		rl.DrawRectangleRoundedLinesEx(barBgRect, 0.4, 6, 1, rl.NewColor(50, 70, 95, 255))

		progStr := fmt.Sprintf("%d/%d", quest.CurrentProgress, quest.TargetProgress)
		rl.DrawText(progStr, int32(barX+barW+12), int32(barY-1), 12, rl.White)

		if quest.IsCompleted() {
			rl.DrawText("✓ COMPLETED", int32(cardX+cardW-120), int32(currY+26), 13, rl.NewColor(0, 220, 130, 255))
		} else {
			rl.DrawText("IN PROGRESS", int32(cardX+cardW-120), int32(currY+26), 12, rl.NewColor(180, 190, 210, 255))
		}

		currY += cardH + gap
	}

	if maxScroll > 0 {
		trackH := listAreaH
		thumbH := (listAreaH / totalContentH) * trackH
		if thumbH < 24 {
			thumbH = 24
		}
		thumbY := listAreaY + (qc.ScrollY/maxScroll)*(trackH-thumbH)
		scrollBarRect := rl.NewRectangle(modalX+modalW-10, thumbY, 4, thumbH)
		rl.DrawRectangleRounded(scrollBarRect, 0.5, 4, rl.NewColor(0, 180, 255, 180))
	}

	rl.EndScissorMode()
}