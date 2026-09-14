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
	MapManager     *service.MapManager
	InventoryMgr   *service.InventoryManager
	ChatMgr        *service.ChatManager
	QuestMgr       *service.QuestManager
	RoomDetailsMgr *service.RoomDetailsManager
	MapComp        *component.MapComponent
	InventoryComp  *component.InventoryComponent
	MainChar       *component.MainCharacter
	HPBar          *component.HPBar
	ChatComp       *component.ChatComponent
	QuestComp      *component.QuestComponent
	RoomDetailsComp *component.RoomDetailsComponent
	PathResolver   *utils.PathResolver
	WindowWidth    int32
	WindowHeight   int32
	Camera         rl.Camera2D
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
	roomDetailsMgr := service.NewRoomDetailsManager()

	// Initial room details test data setup
	roomDetailsMgr.SetRoomDetails("Village Square", []service.RoomItemDetail{
		{ID: "156:sword", Name: "Magic Sword", Description: "Cras mattis consectetur purus sit amet fermentum."},
		{ID: "241:herb", Name: "Healthy potion", Description: "Cras mattis consectetur purus sit amet fermentum."},
	}, []service.RoomNPCDetail{
		{Key: "guard", Name: "Guard knight", Description: "Cras mattis consectetur purus sit amet fermentum.", Kind: "guard", HasQuest: false},
		{Key: "seller", Name: "Seller", Description: "Cras mattis consectetur purus sit amet fermentum.", Kind: "seller", HasQuest: true, QuestID: "npc_2"},
	})

	mapComp := component.NewMapComponent(mapManager)
	inventoryComp := component.NewInventoryComponent(inventoryMgr, pathResolver)
	chatComp := component.NewChatComponent(chatMgr)
	questComp := component.NewQuestComponent(questMgr)
	roomDetailsComp := component.NewRoomDetailsComponent(roomDetailsMgr, pathResolver)

	startX, startY := mapManager.GetStartPosition()

	mainChar := component.NewMainCharacter(startX, startY, service.CharScale, pathResolver.Resolve("characters", "main_character"))
	hpBar := component.NewHPBar(20, 20, 220, 24)

	gameViewWidth := winWidth / 3

	gp := &GamePage{
		WindowWidth:     winWidth,
		WindowHeight:    winHeight,
		PathResolver:    pathResolver,
		MapManager:      mapManager,
		InventoryMgr:    inventoryMgr,
		ChatMgr:         chatMgr,
		QuestMgr:        questMgr,
		RoomDetailsMgr: roomDetailsMgr,
		MapComp:         mapComp,
		InventoryComp:   inventoryComp,
		MainChar:        mainChar,
		HPBar:           hpBar,
		ChatComp:        chatComp,
		QuestComp:       questComp,
		RoomDetailsComp: roomDetailsComp,
		Camera: rl.Camera2D{
			Target:   mainChar.Position,
			Offset:   rl.NewVector2(float32(gameViewWidth)/2, float32(winHeight)/2),
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

	rl.ClearBackground(rl.NewColor(15, 20, 25, 255))

	col1W := gp.WindowWidth / 3
	col2W := gp.WindowWidth - col1W

	// -----------------------------------------------------------------
	// COLUMN 1 (1/3 Width): Game View
	// -----------------------------------------------------------------
	rl.BeginScissorMode(0, 0, col1W, gp.WindowHeight)

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

	// Game View Controls HUD
	hudY := gp.WindowHeight - 65
	rl.DrawRectangle(10, hudY, col1W-20, 55, rl.NewColor(0, 0, 0, 190))
	rl.DrawRectangleLines(10, hudY, col1W-20, 55, rl.NewColor(0, 180, 255, 255))
	rl.DrawText("WASD: Move | SPACE: Attack", 20, hudY+10, 12, rl.White)
	rl.DrawText("I: Bag | Q: Quests", 20, hudY+28, 12, rl.Gold)

	rl.EndScissorMode()

	// Column Divider Border Line
	rl.DrawLineEx(rl.NewVector2(float32(col1W), 0), rl.NewVector2(float32(col1W), float32(gp.WindowHeight)), 2, rl.NewColor(0, 180, 255, 255))

	// -----------------------------------------------------------------
	// COLUMNS 2 & 3 (2/3 Width): Room Details Component
	// -----------------------------------------------------------------
	rightAreaX := float32(col1W)
	rightAreaY := float32(0)
	rightAreaW := float32(col2W)
	rightAreaH := float32(gp.WindowHeight)

	// Draw Background for Right Column Area
	rl.DrawRectangle(int32(rightAreaX), 0, col2W, gp.WindowHeight, rl.NewColor(220, 225, 230, 255))

	if gp.RoomDetailsComp != nil {
		gp.RoomDetailsComp.Render(rightAreaX, rightAreaY, rightAreaW, rightAreaH)
	}

	// Overlays
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