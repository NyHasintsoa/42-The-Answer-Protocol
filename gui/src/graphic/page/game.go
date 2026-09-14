package page

import (
	"fmt"

	"tap-gui/src/graphic/component"
	"tap-gui/src/model/enums"
	"tap-gui/src/service"
	"tap-gui/src/utils"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type GamePage struct {
	BasePage
	MapManager    *service.MapManager
	InventoryMgr  *service.InventoryManager
	ChatMgr       *service.ChatManager
	QuestMgr      *service.QuestManager
	MapComp       *component.MapComponent
	InventoryComp *component.InventoryComponent
	MainChar      *component.MainCharacter
	HPBar         *component.HPBar
	ChatComp      *component.ChatComponent
	QuestComp     *component.QuestComponent
	PathResolver  *utils.PathResolver
	WindowWidth   int32
	WindowHeight  int32
	Camera        rl.Camera2D
}

func NewGamePage(winWidth, winHeight int32) *GamePage {
	pathResolver := utils.NewPathResolver("assets")
	mapManager := service.NewMapManager(pathResolver, "", service.EntityFactory{
		NewMonster: func(x, y, scale float32, assetPath, name string) service.Monster {
			return component.NewMonster(x, y, scale, assetPath, name)
		},
		NewNPC: func(kind, relativeDir string, x, y, scale float32) (service.INpc, error) {
			switch kind {
			case "guard":
				return component.NewGuardCharacter(pathResolver, relativeDir, x, y, scale)
			case "seller":
				return component.NewSellerCharacter(pathResolver, relativeDir, x, y, scale)
			default:
				return component.NewPersonCharacter(pathResolver, relativeDir, x, y, scale)
			}
		},
	})
	inventoryMgr := service.NewInventoryManager()
	chatMgr := service.NewChatManager()
	questMgr := service.NewQuestManager()

	mapComp := component.NewMapComponent(mapManager)
	inventoryComp := component.NewInventoryComponent(inventoryMgr, pathResolver)
	chatComp := component.NewChatComponent(chatMgr)
	questComp := component.NewQuestComponent(questMgr)

	startX, startY := mapManager.GetStartPosition()

	mainChar := component.NewMainCharacter(startX, startY, service.CharScale, pathResolver.Resolve("characters", "main_character"))
	hpBar := component.NewHPBar(30, 20, 320, 28)

	gp := &GamePage{
		WindowWidth:   winWidth,
		WindowHeight:  winHeight,
		PathResolver:  pathResolver,
		MapManager:    mapManager,
		InventoryMgr:  inventoryMgr,
		ChatMgr:       chatMgr,
		QuestMgr:      questMgr,
		MapComp:       mapComp,
		InventoryComp: inventoryComp,
		MainChar:      mainChar,
		HPBar:         hpBar,
		ChatComp:      chatComp,
		QuestComp:     questComp,
		Camera: rl.Camera2D{
			Target:   mainChar.Position,
			Offset:   rl.NewVector2(float32(winWidth)/2, float32(winHeight)/2),
			Rotation: 0.0,
			Zoom:     1.0,
		},
	}
	gp.State = enums.GamePage

	return gp
}

func (gp *GamePage) Unload() {
	if gp.IsUnloaded {
		return
	}
	if gp.MapManager != nil {
		gp.MapManager.Unload()
	}
	if gp.MainChar != nil {
		gp.MainChar.UnloadAnimations()
	}
	gp.BasePage.Unload()
}

func (gp *GamePage) EventListener() {
	if rl.IsKeyPressed(rl.KeyEscape) {
		gp.NextState = enums.MainMenu
	}

	if rl.IsKeyPressed(rl.KeyQ) && gp.QuestMgr != nil {
		gp.QuestMgr.Toggle()
	}

	if rl.IsKeyPressed(rl.KeyMinus) || rl.IsKeyPressed(rl.KeyKpSubtract) {
		gp.HPBar.SetHP(gp.HPBar.TargetHP - 20)
	}
	if rl.IsKeyPressed(rl.KeyEqual) || rl.IsKeyPressed(rl.KeyKpAdd) {
		gp.HPBar.SetHP(gp.HPBar.TargetHP + 20)
	}
}

func (gp *GamePage) Update(dt float32) {
	if gp.InventoryMgr != nil {
		gp.InventoryMgr.Update(gp.WindowWidth, gp.WindowHeight)
	}

	if gp.MainChar.IsDead {
		gp.MainChar.AdvanceFrame(dt, false)
		return
	}

	if (gp.InventoryMgr != nil && gp.InventoryMgr.IsOpen) ||
		(gp.ChatMgr != nil && gp.ChatMgr.IsOpen) ||
		(gp.QuestMgr != nil && gp.QuestMgr.IsOpen) {
		gp.MainChar.SetAction(component.ActionIdle)
		return
	}

	if gp.MainChar.Action == component.ActionAttacking || gp.MainChar.Action == component.ActionHurt {
		if gp.MainChar.AdvanceFrame(dt, false) {
			gp.MainChar.SetAction(component.ActionIdle)
		}
		return
	}

	if rl.IsKeyPressed(rl.KeySpace) {
		gp.MainChar.SetAction(component.ActionAttacking)
		return
	}

	moveDir := rl.NewVector2(0, 0)
	newDir := gp.MainChar.Direction

	if rl.IsKeyDown(rl.KeyUp) || rl.IsKeyDown(rl.KeyW) {
		moveDir.Y -= 1
		newDir = component.DirBack
	}
	if rl.IsKeyDown(rl.KeyDown) || rl.IsKeyDown(rl.KeyS) {
		moveDir.Y += 1
		newDir = component.DirFront
	}
	if rl.IsKeyDown(rl.KeyLeft) || rl.IsKeyDown(rl.KeyA) {
		moveDir.X -= 1
		newDir = component.DirLeft
	}
	if rl.IsKeyDown(rl.KeyRight) || rl.IsKeyDown(rl.KeyD) {
		moveDir.X += 1
		newDir = component.DirRight
	}

	if newDir != gp.MainChar.Direction {
		gp.MainChar.Direction = newDir
		gp.MainChar.FrameIndex = 0
		gp.MainChar.FrameTimer = 0
	}

	if moveDir.X != 0 || moveDir.Y != 0 {
		gp.MainChar.SetAction(component.ActionWalking)
		length := float32(rl.Vector2Length(moveDir))

		nextX := gp.MainChar.Position.X + (moveDir.X/length)*gp.MainChar.Speed*dt
		nextY := gp.MainChar.Position.Y + (moveDir.Y/length)*gp.MainChar.Speed*dt

		radius := float32(8.0)
		if gp.MapManager.IsWalkable(nextX-radius, gp.MainChar.Position.Y-radius) &&
			gp.MapManager.IsWalkable(nextX+radius, gp.MainChar.Position.Y+radius) {
			gp.MainChar.Position.X = nextX
		}
		if gp.MapManager.IsWalkable(gp.MainChar.Position.X-radius, nextY-radius) &&
			gp.MapManager.IsWalkable(gp.MainChar.Position.X+radius, nextY+radius) {
			gp.MainChar.Position.Y = nextY
		}
	} else {
		gp.MainChar.SetAction(component.ActionIdle)
	}

	gp.MapManager.Update(dt, gp.MainChar.Position)
	gp.Camera.Target = gp.MainChar.Position
	gp.MainChar.AdvanceFrame(dt, true)

	if gp.HPBar != nil {
		gp.HPBar.Update(dt)
	}
}

func (gp *GamePage) Render() {
	gp.EventListener()
	dt := rl.GetFrameTime()
	gp.Update(dt)

	rl.ClearBackground(rl.NewColor(15, 20, 15, 255))

	rl.BeginMode2D(gp.Camera)
	if gp.MapComp != nil {
		gp.MapComp.Render()
	}
	if gp.MainChar != nil {
		gp.MainChar.Render()
	}
	rl.EndMode2D()

	if gp.HPBar != nil {
		gp.HPBar.Render()
	}

	rl.DrawRectangle(10, gp.WindowHeight-70, 520, 60, rl.NewColor(0, 0, 0, 180))
	rl.DrawRectangleLines(10, gp.WindowHeight-70, 520, 60, rl.RayWhite)
	rl.DrawText("WASD: Move | SPACE: Attack | I: Inventory | C: Chat | Q: Quests | [- / +]: Test HP", 20, gp.WindowHeight-62, 12, rl.RayWhite)

	if gp.MapManager.ActiveRoom != nil {
		infoText := fmt.Sprintf("Room: %s", gp.MapManager.ActiveRoom.Name)
		rl.DrawText(infoText, 20, gp.WindowHeight-44, 14, rl.Gold)
	}

	qBtnW := float32(110)
	qBtnH := float32(40)
	qBtnX := float32(gp.WindowWidth) - qBtnW - 20
	qBtnY := float32(gp.WindowHeight) - qBtnH - 20
	qBtnRect := rl.NewRectangle(qBtnX, qBtnY, qBtnW, qBtnH)

	mousePos := rl.GetMousePosition()
	qBtnBg := rl.NewColor(18, 22, 34, 230)
	qBtnBorder := rl.NewColor(0, 180, 255, 255)

	if rl.CheckCollisionPointRec(mousePos, qBtnRect) {
		qBtnBg = rl.NewColor(30, 50, 80, 255)
		if rl.IsMouseButtonPressed(rl.MouseLeftButton) && gp.QuestMgr != nil {
			gp.QuestMgr.Toggle()
		}
	}

	rl.DrawRectangleRounded(qBtnRect, 0.25, 8, qBtnBg)
	rl.DrawRectangleRoundedLinesEx(qBtnRect, 0.25, 8, 2, qBtnBorder)
	rl.DrawText("QUESTS [Q]", int32(qBtnX+12), int32(qBtnY+12), 15, rl.White)

	if gp.InventoryComp != nil {
		gp.InventoryComp.Render(gp.WindowWidth, gp.WindowHeight)
	}

	if gp.ChatComp != nil {
		gp.ChatComp.Render(gp.WindowWidth, gp.WindowHeight)
	}

	if gp.QuestComp != nil {
		gp.QuestComp.Render(gp.WindowWidth, gp.WindowHeight)
	}
}