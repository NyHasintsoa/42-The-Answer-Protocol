package component

import (
	"strings"

	"tap-gui/src/service"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type ChatComponent struct {
	Manager        *service.ChatManager
	InputComp      *InputComponent
	ScrollY        float32
	scrollToBottom bool
}

func NewChatComponent(mgr *service.ChatManager) *ChatComponent {
	return &ChatComponent{
		Manager:        mgr,
		InputComp:      NewInputComponent(0, 0, 0, 0, "Type a message...", 140),
		ScrollY:        0,
		scrollToBottom: true,
	}
}

func wrapText(text string, fontSize int32, maxWidth float32) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	var lines []string
	currentLine := words[0]

	for _, word := range words[1:] {
		testLine := currentLine + " " + word
		if float32(rl.MeasureText(testLine, fontSize)) <= maxWidth {
			currentLine = testLine
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	lines = append(lines, currentLine)
	return lines
}

func (cc *ChatComponent) Render(winWidth, winHeight int32) {
	if cc.Manager == nil {
		return
	}

	dt := rl.GetFrameTime()
	mousePos := rl.GetMousePosition()
	mouseClicked := rl.IsMouseButtonPressed(rl.MouseLeftButton)

	
	btnW := float32(110)
	btnH := float32(40)
	btnX := float32(20)
	btnY := (float32(winHeight) - btnH) / 2
	btnRect := rl.NewRectangle(btnX, btnY, btnW, btnH)

	btnBg := rl.NewColor(30, 35, 55, 240)
	btnBorder := rl.NewColor(0, 180, 255, 255)
	if rl.CheckCollisionPointRec(mousePos, btnRect) {
		btnBg = rl.NewColor(50, 70, 110, 255)
		btnBorder = rl.Gold
		if mouseClicked {
			cc.Manager.Toggle()
			if cc.Manager.IsOpen {
				cc.scrollToBottom = true
			}
		}
	}

	rl.DrawRectangleRounded(btnRect, 0.25, 8, btnBg)
	rl.DrawRectangleRoundedLinesEx(btnRect, 0.25, 8, 2, btnBorder)
	rl.DrawText("CHAT [C]", int32(btnX+20), int32(btnY+11), 18, rl.RayWhite)

	if rl.IsKeyPressed(rl.KeyC) {
		cc.Manager.Toggle()
		if cc.Manager.IsOpen {
			cc.scrollToBottom = true
		}
	}

	if !cc.Manager.IsOpen {
		return
	}

	
	rl.DrawRectangle(0, 0, winWidth, winHeight, rl.NewColor(0, 0, 0, 160))

	
	modalW := float32(760)
	modalH := float32(680)
	modalX := (float32(winWidth) - modalW) / 2
	modalY := (float32(winHeight) - modalH) / 2
	modalRect := rl.NewRectangle(modalX, modalY, modalW, modalH)

	rl.DrawRectangleRounded(modalRect, 0.05, 12, rl.NewColor(18, 22, 34, 250))
	rl.DrawRectangleRoundedLinesEx(modalRect, 0.05, 12, 2, rl.NewColor(0, 180, 255, 255))

	
	headerH := float32(60)
	headerRect := rl.NewRectangle(modalX, modalY, modalW, headerH)
	rl.DrawRectangleRounded(headerRect, 0.08, 12, rl.NewColor(14, 18, 28, 255))
	rl.DrawLineEx(rl.NewVector2(modalX, modalY+headerH), rl.NewVector2(modalX+modalW, modalY+headerH), 2, rl.NewColor(0, 180, 255, 255))

	tabs := []service.ChatChannel{service.ChannelGlobal, service.ChannelRoom, service.ChannelGroup}
	tabStartX := modalX + 24
	tabY := modalY + 14

	for _, tab := range tabs {
		tabText := tab.String()
		tabW := float32(rl.MeasureText(tabText, 16) + 30)
		tabRect := rl.NewRectangle(tabStartX, tabY, tabW, 32)

		isActive := cc.Manager.ActiveTab == tab
		tabBg := rl.NewColor(30, 38, 55, 255)
		borderColor := rl.NewColor(50, 65, 90, 255)
		textColor := rl.NewColor(160, 175, 200, 255)

		if isActive {
			tabBg = rl.NewColor(0, 130, 230, 255)
			borderColor = rl.NewColor(0, 200, 255, 255)
			textColor = rl.White
		} else if rl.CheckCollisionPointRec(mousePos, tabRect) {
			tabBg = rl.NewColor(45, 58, 85, 255)
			textColor = rl.RayWhite
			if mouseClicked {
				cc.Manager.ActiveTab = tab
				cc.scrollToBottom = true
			}
		}

		rl.DrawRectangleRounded(tabRect, 0.25, 8, tabBg)
		rl.DrawRectangleRoundedLinesEx(tabRect, 0.25, 8, 1, borderColor)
		rl.DrawText(tabText, int32(tabStartX+15), int32(tabY+7), 16, textColor)

		tabStartX += tabW + 12
	}

	
	closeBtn := rl.NewRectangle(modalX+modalW-42, modalY+14, 30, 30)
	closeCol := rl.NewColor(180, 50, 50, 255)
	if rl.CheckCollisionPointRec(mousePos, closeBtn) {
		closeCol = rl.Red
		if mouseClicked {
			cc.Manager.IsOpen = false
		}
	}
	rl.DrawRectangleRounded(closeBtn, 0.3, 8, closeCol)
	rl.DrawRectangleRoundedLinesEx(closeBtn, 0.3, 8, 1, rl.White)
	rl.DrawText("X", int32(modalX+modalW-33), int32(modalY+19), 18, rl.White)

	
	messages := cc.Manager.GetCurrentMessages()
	msgAreaY := modalY + headerH + 15
	msgAreaH := modalH - headerH - 90
	msgAreaRect := rl.NewRectangle(modalX, msgAreaY, modalW, msgAreaH)

	fontSize := int32(14)
	paddingX := float32(18)
	paddingY := float32(12)
	lineHeight := float32(18)
	maxTextWidth := modalW - 280

	
	var totalContentH float32 = 20
	type messageLayout struct {
		lines   []string
		bubbleW float32
		bubbleH float32
	}
	layouts := make([]messageLayout, len(messages))

	for i, msg := range messages {
		lines := wrapText(msg.Text, fontSize, maxTextWidth)
		var maxLineW float32 = 0
		for _, l := range lines {
			w := float32(rl.MeasureText(l, fontSize))
			if w > maxLineW {
				maxLineW = w
			}
		}
		bW := maxLineW + (paddingX * 2)
		bH := float32(len(lines))*lineHeight + (paddingY * 2)

		layouts[i] = messageLayout{lines: lines, bubbleW: bW, bubbleH: bH}
		totalContentH += bH + 20
	}

	maxScroll := totalContentH - msgAreaH
	if maxScroll < 0 {
		maxScroll = 0
	}

	
	if rl.CheckCollisionPointRec(mousePos, msgAreaRect) {
		wheel := rl.GetMouseWheelMove()
		if wheel != 0 {
			cc.ScrollY -= wheel * 35
		}
	}

	
	if cc.scrollToBottom {
		cc.ScrollY = maxScroll
		cc.scrollToBottom = false
	}

	
	if cc.ScrollY < 0 {
		cc.ScrollY = 0
	}
	if cc.ScrollY > maxScroll {
		cc.ScrollY = maxScroll
	}

	
	rl.BeginScissorMode(int32(modalX), int32(msgAreaY), int32(modalW), int32(msgAreaH))

	currY := msgAreaY + 10 - cc.ScrollY
	avatarRadius := float32(22)

	for i, msg := range messages {
		layout := layouts[i]
		bubbleW := layout.bubbleW
		bubbleH := layout.bubbleH
		lines := layout.lines

		if msg.IsSelf {
			avatarX := modalX + modalW - 40
			avatarY := currY + avatarRadius

			rl.DrawCircleV(rl.NewVector2(avatarX, avatarY), avatarRadius, msg.AvatarCol)
			rl.DrawText("Y", int32(avatarX-5), int32(avatarY-8), 16, rl.White)

			bubbleX := avatarX - avatarRadius - 14 - bubbleW
			bubbleY := currY
			bubbleRect := rl.NewRectangle(bubbleX, bubbleY, bubbleW, bubbleH)

			rl.DrawRectangleRounded(bubbleRect, 0.25, 8, rl.NewColor(0, 110, 200, 255))
			rl.DrawRectangleRoundedLinesEx(bubbleRect, 0.25, 8, 1, rl.NewColor(0, 180, 255, 255))
			rl.DrawTriangle(
				rl.NewVector2(bubbleX+bubbleW, bubbleY+14),
				rl.NewVector2(bubbleX+bubbleW, bubbleY+26),
				rl.NewVector2(bubbleX+bubbleW+8, bubbleY+20),
				rl.NewColor(0, 110, 200, 255),
			)

			for idx, line := range lines {
				lineY := bubbleY + paddingY + float32(idx)*lineHeight
				rl.DrawText(line, int32(bubbleX+paddingX), int32(lineY), fontSize, rl.White)
			}

			statusW := float32(rl.MeasureText(msg.Status, 11))
			rl.DrawText(msg.Status, int32(bubbleX-statusW-12), int32(bubbleY+paddingY), 11, rl.NewColor(130, 150, 180, 255))

		} else {
			avatarX := modalX + 40
			avatarY := currY + avatarRadius

			rl.DrawCircleV(rl.NewVector2(avatarX, avatarY), avatarRadius, msg.AvatarCol)
			initial := "M"
			if len(msg.Sender) > 0 {
				initial = string(msg.Sender[0])
			}
			rl.DrawText(initial, int32(avatarX-5), int32(avatarY-8), 16, rl.White)

			bubbleX := avatarX + avatarRadius + 14
			bubbleY := currY
			bubbleRect := rl.NewRectangle(bubbleX, bubbleY, bubbleW, bubbleH)

			rl.DrawRectangleRounded(bubbleRect, 0.25, 8, rl.NewColor(32, 42, 62, 255))
			rl.DrawRectangleRoundedLinesEx(bubbleRect, 0.25, 8, 1, rl.NewColor(0, 160, 220, 255))
			rl.DrawTriangle(
				rl.NewVector2(bubbleX, bubbleY+26),
				rl.NewVector2(bubbleX, bubbleY+14),
				rl.NewVector2(bubbleX-8, bubbleY+20),
				rl.NewColor(32, 42, 62, 255),
			)

			for idx, line := range lines {
				lineY := bubbleY + paddingY + float32(idx)*lineHeight
				rl.DrawText(line, int32(bubbleX+paddingX), int32(lineY), fontSize, rl.RayWhite)
			}

			rl.DrawText(msg.Status, int32(bubbleX+bubbleW+12), int32(bubbleY+paddingY), 11, rl.NewColor(130, 150, 180, 255))
		}

		currY += bubbleH + 20
	}

	
	if maxScroll > 0 {
		trackH := msgAreaH
		thumbH := (msgAreaH / totalContentH) * trackH
		if thumbH < 24 {
			thumbH = 24
		}
		thumbY := msgAreaY + (cc.ScrollY/maxScroll)*(trackH-thumbH)
		scrollBarRect := rl.NewRectangle(modalX+modalW-10, thumbY, 5, thumbH)
		rl.DrawRectangleRounded(scrollBarRect, 0.5, 4, rl.NewColor(0, 180, 255, 180))
	}

	rl.EndScissorMode()

	
	inputMargin := float32(20)
	inputW := modalW - (inputMargin * 2)
	inputH := float32(48)
	inputX := modalX + inputMargin
	inputY := modalY + modalH - inputH - 18

	cc.InputComp.Rect = rl.NewRectangle(inputX, inputY, inputW, inputH)

	submitted := cc.InputComp.Update(dt)
	cc.InputComp.Render()

	if submitted {
		text := cc.InputComp.GetAndClearText()
		if strings.TrimSpace(text) != "" {
			cc.Manager.SendMessage(text)
			cc.scrollToBottom = true
		}
	}
}