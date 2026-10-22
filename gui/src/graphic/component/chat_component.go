package component

import (
	"strings"

	"tap-gui/src/graphic/component/ui"
	"tap-gui/src/service"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type ChatComponent struct {
	Manager        *service.ChatManager
	InputComp      *ui.InputComponent
	ScrollY        float32
	scrollToBottom bool
	OnLog          func(string)
}

// NewChatComponent creates a chat view backed by mgr.
func NewChatComponent(mgr *service.ChatManager) *ChatComponent {
	return &ChatComponent{
		Manager:        mgr,
		InputComp:      ui.NewInputComponent(0, 0, 0, 48, "Type a message...", 280),
		ScrollY:        0,
		scrollToBottom: true,
	}
}

// Render draws the chat interface within the specified window dimensions.
// winWidth and winHeight are the window dimensions in pixels.
func (cc *ChatComponent) Render(winWidth, winHeight int32) {
	if cc.Manager == nil {
		return
	}

	dt := rl.GetFrameTime()

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

		isActive := cc.Manager.ActiveTab == tab
		tabButton := ui.NewButton(tabStartX, tabY, tabW, 32, tabText, 16)
		tabButton.BgColor = rl.NewColor(30, 38, 55, 255)
		tabButton.HoverBgColor = rl.NewColor(45, 58, 85, 255)
		tabButton.ClickedColor = tabButton.HoverBgColor
		tabButton.TextColor = rl.NewColor(160, 175, 200, 255)
		tabButton.BorderColor = rl.NewColor(50, 65, 90, 255)
		tabButton.ShadowColor = rl.Blank
		tabButton.BorderWidth = 1
		tabButton.BorderRadius = 0.25
		if isActive {
			tabButton.BgColor = rl.NewColor(0, 130, 230, 255)
			tabButton.HoverBgColor = tabButton.BgColor
			tabButton.ClickedColor = tabButton.BgColor
			tabButton.TextColor = rl.White
			tabButton.BorderColor = rl.NewColor(0, 200, 255, 255)
		}
		tabButton.Render()
		if tabButton.IsClicked && !isActive {
			cc.Manager.ActiveTab = tab
			cc.scrollToBottom = true
			if cc.OnLog != nil {
				cc.OnLog("Switched to " + tabText + " chat.")
			}
		}

		tabStartX += tabW + 12
	}

	closeButton := ui.NewButton(modalX+modalW-42, modalY+14, 30, 30, "X", 18)
	closeButton.BgColor = rl.NewColor(180, 50, 50, 255)
	closeButton.HoverBgColor = rl.Red
	closeButton.ClickedColor = closeButton.HoverBgColor
	closeButton.TextColor = rl.White
	closeButton.BorderColor = rl.White
	closeButton.ShadowColor = rl.Blank
	closeButton.BorderWidth = 1
	closeButton.BorderRadius = 0.3
	closeButton.Render()
	if closeButton.IsClicked {
		cc.Manager.IsOpen = false
		if cc.OnLog != nil {
			cc.OnLog("Closed chat.")
		}
	}

	// Update Textarea Input Component layout dynamically
	inputMargin := float32(20)
	inputW := modalW - (inputMargin * 2)
	cc.InputComp.Rect.Width = inputW

	submitted := cc.InputComp.Update(dt)

	inputH := cc.InputComp.Rect.Height
	inputX := modalX + inputMargin
	inputY := modalY + modalH - inputH - 18
	cc.InputComp.Rect.X = inputX
	cc.InputComp.Rect.Y = inputY

	if submitted {
		text := cc.InputComp.GetAndClearText()
		if strings.TrimSpace(text) != "" {
			cc.Manager.SendMessage(text)
			if cc.OnLog != nil {
				cc.OnLog("Sent a message in " + cc.Manager.ActiveTab.String() + " chat.")
			}
			cc.scrollToBottom = true
		}
	}

	messages := cc.Manager.GetCurrentMessages()
	msgAreaY := modalY + headerH + 15
	msgAreaH := modalH - headerH - inputH - 45
	msgAreaRect := rl.NewRectangle(modalX, msgAreaY, modalW, msgAreaH)

	fontSize := int32(14)
	paddingX := float32(18)
	paddingY := float32(12)
	maxTextWidth := int32(modalW - 280)

	var totalContentH float32 = 20
	type messageLayout struct {
		bubbleW float32
		bubbleH float32
	}
	layouts := make([]messageLayout, len(messages))

	for i, msg := range messages {
		lines := ui.WrapTextLines(msg.Text, fontSize, maxTextWidth)
		var maxLineW float32 = 0
		for _, l := range lines {
			w := float32(rl.MeasureText(l, fontSize))
			if w > maxLineW {
				maxLineW = w
			}
		}
		bW := maxLineW + (paddingX * 2)
		bH := float32(len(lines))*float32(fontSize+2) + (paddingY * 2)

		layouts[i] = messageLayout{bubbleW: bW, bubbleH: bH}
		totalContentH += bH + 20
	}

	maxScroll := totalContentH - msgAreaH
	if maxScroll < 0 {
		maxScroll = 0
	}

	if cc.scrollToBottom {
		cc.ScrollY = maxScroll
		cc.scrollToBottom = false
	}
	ui.UpdateScrollOffset(&cc.ScrollY, msgAreaRect, totalContentH)

	rl.BeginScissorMode(int32(modalX), int32(msgAreaY), int32(modalW), int32(msgAreaH))

	currY := msgAreaY + 10 - cc.ScrollY
	avatarRadius := float32(22)

	for i, msg := range messages {
		layout := layouts[i]
		bubbleW := layout.bubbleW
		bubbleH := layout.bubbleH

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

			ui.DrawTextWrapped(msg.Text, int32(bubbleX+paddingX), int32(bubbleY+paddingY), maxTextWidth, fontSize, rl.White)

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

			ui.DrawTextWrapped(msg.Text, int32(bubbleX+paddingX), int32(bubbleY+paddingY), maxTextWidth, fontSize, rl.RayWhite)

			rl.DrawText(msg.Status, int32(bubbleX+bubbleW+12), int32(bubbleY+paddingY), 11, rl.NewColor(130, 150, 180, 255))
		}

		currY += bubbleH + 20
	}

	rl.EndScissorMode()
	ui.DrawScrollIndicator(msgAreaRect, totalContentH, cc.ScrollY)

	cc.InputComp.Render()
}
