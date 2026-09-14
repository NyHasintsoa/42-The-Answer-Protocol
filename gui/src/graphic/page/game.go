package page

import (
	"tap-gui/src/graphic/component"
	"tap-gui/src/model/enums"
	"tap-gui/src/service"
	"tap-gui/src/utils"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type GamePage struct {
	BasePage
	MapManager      *service.MapManager
	InventoryMgr    *service.InventoryManager
	ChatMgr         *service.ChatManager
	QuestMgr        *service.QuestManager
	RoomDetailsMgr  *service.RoomDetailsManager
	RoomViewMgr     *component.RoomViewManager
	PlayerMgr       *component.PlayerManager
	MapComp         *component.MapComponent
	MainChar        *component.MainCharacter
	HPBar           *component.HPBar
	InventoryComp   *component.InventoryComponent
	ChatComp        *component.ChatComponent
	QuestComp       *component.QuestComponent
	RoomDetailsComp *component.RoomDetailsComponent
	RoomViewComp    *component.RoomViewComponent
	PlayerComp      *component.PlayerComponent
	ActionBtnsComp  *component.ActionButtonsComponent
	LogViewComp     *component.LogViewComponent
	PathResolver    *utils.PathResolver
	WindowWidth     int32
	WindowHeight    int32
	Camera          rl.Camera2D
}

func NewGamePage(winWidth, winHeight int32) *GamePage {
	pathResolver := utils.NewPathResolver("assets")
	mapManager := service.NewMapManager(pathResolver, "", service.EntityFactory{
		NewMonster: func(x, y, scale float32, resolver *utils.PathResolver, assetPath, name string) service.Monster {
			return component.NewMonster(x, y, scale, resolver, assetPath, name)
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

	questMgr := service.NewQuestManager()
	inventoryMgr := service.NewInventoryManager()
	chatMgr := service.NewChatManager()
	roomDetailsMgr := service.NewRoomDetailsManager()
	roomViewMgr := component.NewRoomViewManager()
	playerMgr := component.NewPlayerManager()

	roomDetailsMgr.SetRoomDetails("Village Square", []service.RoomItemDetail{
		{ID: "156:sword", Name: "Magic Sword", Description: "Cras mattis consectetur purus sit amet fermentum."},
		{ID: "241:potion_health_PM", Name: "Healthy potion", Description: "Cras mattis consectetur purus sit amet fermentum."},
	}, []service.RoomNPCDetail{
		{Key: "guard", Name: "Guard knight", Description: "Cras mattis consectetur purus sit amet fermentum.", Kind: "guard", HasQuest: false},
		{Key: "seller", Name: "Seller", Description: "Cras mattis consectetur purus sit amet fermentum.", Kind: "seller", HasQuest: true, QuestID: "npc_2"},
	})

	mapComp := component.NewMapComponent(mapManager)
	inventoryComp := component.NewInventoryComponent(inventoryMgr, pathResolver)
	chatComp := component.NewChatComponent(chatMgr)
	questComp := component.NewQuestComponent(questMgr)
	roomDetailsComp := component.NewRoomDetailsComponent(roomDetailsMgr, pathResolver)
	roomViewComp := component.NewRoomViewComponent(roomViewMgr)
	playerComp := component.NewPlayerComponent(playerMgr)
	actionBtnsComp := component.NewActionButtonsComponent(inventoryMgr, chatMgr, questMgr)
	logViewComp := component.NewLogViewComponent()

	startX, startY := mapManager.GetStartPosition()

	mainChar := component.NewMainCharacter(startX, startY, service.CharScale, pathResolver, pathResolver.Resolve("characters", "main_character"))
	hpBar := component.NewHPBar(15, 15, 180, 20)

	mapViewSize := float32(winWidth) / 3
	if mapViewSize > float32(winHeight)/2 {
		mapViewSize = float32(winHeight) / 2
	}

	gp := &GamePage{
		WindowWidth:     winWidth,
		WindowHeight:    winHeight,
		PathResolver:    pathResolver,
		MapManager:      mapManager,
		InventoryMgr:    inventoryMgr,
		ChatMgr:         chatMgr,
		QuestMgr:        questMgr,
		RoomDetailsMgr:  roomDetailsMgr,
		RoomViewMgr:     roomViewMgr,
		PlayerMgr:       playerMgr,
		MapComp:         mapComp,
		MainChar:        mainChar,
		HPBar:           hpBar,
		InventoryComp:   inventoryComp,
		ChatComp:        chatComp,
		QuestComp:       questComp,
		RoomDetailsComp: roomDetailsComp,
		RoomViewComp:    roomViewComp,
		PlayerComp:      playerComp,
		ActionBtnsComp:  actionBtnsComp,
		LogViewComp:     logViewComp,
		Camera: rl.Camera2D{
			Target:   mainChar.Position,
			Offset:   rl.NewVector2(mapViewSize/2, mapViewSize/2),
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
	if gp.RoomDetailsComp != nil {
		gp.RoomDetailsComp.Unload()
	}
	if gp.InventoryComp != nil {
		gp.InventoryComp.Unload()
	}
	if gp.PathResolver != nil && gp.PathResolver.ImageManager != nil {
		gp.PathResolver.ImageManager.UnloadAll()
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
	if gp.MainChar.IsDead {
		gp.MainChar.AdvanceFrame(dt, false)
		return
	}

	if (gp.QuestMgr != nil && gp.QuestMgr.IsOpen) ||
		(gp.InventoryMgr != nil && gp.InventoryMgr.IsOpen) ||
		(gp.ChatMgr != nil && gp.ChatMgr.IsOpen) {
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

	rl.ClearBackground(rl.NewColor(225, 230, 235, 255))

	totalW := float32(gp.WindowWidth)
	topMapSize := totalW / 3
	maxTopHeight := float32(gp.WindowHeight) / 2
	if topMapSize > maxTopHeight {
		topMapSize = maxTopHeight
	}
	topDetailsW := totalW - topMapSize

	rl.BeginScissorMode(0, 0, int32(topMapSize), int32(topMapSize))
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
	rl.EndScissorMode()
	rl.DrawRectangleLinesEx(rl.NewRectangle(0, 0, topMapSize, topMapSize), 2, rl.NewColor(180, 190, 200, 255))

	if gp.RoomDetailsComp != nil {
		gp.RoomDetailsComp.Render(topMapSize, 0, topDetailsW, topMapSize)
	}

	rl.DrawLineEx(rl.NewVector2(0, topMapSize), rl.NewVector2(totalW, topMapSize), 2, rl.NewColor(180, 190, 200, 255))

	bottomLeftW := totalW * 0.52
	bottomRightW := totalW - bottomLeftW

	botY := topMapSize + 8
	botH := float32(gp.WindowHeight) - topMapSize - 16

	roomPlayerH := botH*0.55 - 4
	roomViewW := (bottomLeftW - 20) * 0.58
	playerRoomW := (bottomLeftW - 20) * 0.42

	roomViewRect := rl.NewRectangle(10, botY, roomViewW, roomPlayerH)
	if gp.RoomViewComp != nil {
		gp.RoomViewComp.Render(roomViewRect)
	}

	playerRoomRect := rl.NewRectangle(10+roomViewW+8, botY, playerRoomW, roomPlayerH)
	if gp.PlayerComp != nil {
		gp.PlayerComp.Render(playerRoomRect)
	}

	actionY := botY + roomPlayerH + 8
	actionH := botH - roomPlayerH - 8
	actionRect := rl.NewRectangle(10, actionY, bottomLeftW-12, actionH)

	if gp.ActionBtnsComp != nil {
		gp.ActionBtnsComp.Render(actionRect)
	}

	logRect := rl.NewRectangle(bottomLeftW+6, botY, bottomRightW-16, botH)
	if gp.LogViewComp != nil {
		gp.LogViewComp.Render(logRect)
	}

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
