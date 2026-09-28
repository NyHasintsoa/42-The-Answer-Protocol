package page

import (
	"strings"

	"tap-gui/src/graphic/component"
	"tap-gui/src/model"
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
	RoomTarget      rl.Vector2
	IsMovingToRoom  bool
}

func NewGamePage(winWidth, winHeight int32, ctx *model.GameContext) *GamePage {
	var resourceManager *utils.ResourceManager
	if ctx != nil {
		resourceManager = ctx.ResourceManager
	}
	pathResolver := utils.NewPathResolverWithResourceManager("assets", resourceManager)
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
	hpBar := component.NewHPBar(30, 15, 180, 20)

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
	gp.RoomViewMgr.OnMove = gp.moveToRoom
	gp.syncRoomContext()

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

func (gp *GamePage) hasOpenModal() bool {
	return (gp.QuestMgr != nil && gp.QuestMgr.IsOpen) ||
		(gp.InventoryMgr != nil && gp.InventoryMgr.IsOpen) ||
		(gp.ChatMgr != nil && gp.ChatMgr.IsOpen)
}

func (gp *GamePage) Update(dt float32) {
	if gp.MainChar.IsDead {
		gp.MainChar.AdvanceFrame(dt, false)
		return
	}

	if gp.hasOpenModal() {
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

	if gp.IsMovingToRoom {
		delta := rl.Vector2Subtract(gp.RoomTarget, gp.MainChar.Position)
		distance := float32(rl.Vector2Length(delta))
		step := gp.MainChar.Speed * dt
		if distance <= step || distance == 0 {
			gp.MainChar.Position = gp.RoomTarget
			gp.IsMovingToRoom = false
			gp.RoomViewMgr.IsMoving = false
			gp.MainChar.SetAction(component.ActionIdle)
		} else {
			moveX := delta.X / distance
			moveY := delta.Y / distance
			gp.MainChar.Position.X += moveX * step
			gp.MainChar.Position.Y += moveY * step
			if moveX < 0 {
				gp.MainChar.Direction = component.DirLeft
			} else if moveX > 0 {
				gp.MainChar.Direction = component.DirRight
			} else if moveY < 0 {
				gp.MainChar.Direction = component.DirBack
			} else {
				gp.MainChar.Direction = component.DirFront
			}
			gp.MainChar.SetAction(component.ActionWalking)
		}
	} else {
		gp.MainChar.SetAction(component.ActionIdle)
	}

	previousRoom := gp.MapManager.ActiveRoom
	gp.MapManager.Update(dt, gp.MainChar.Position)
	if gp.MapManager.ActiveRoom != previousRoom {
		gp.syncRoomContext()
	}
	gp.Camera.Target = gp.MainChar.Position
	gp.MainChar.AdvanceFrame(dt, true)

	if gp.HPBar != nil {
		gp.HPBar.Update(dt)
	}
}

func (gp *GamePage) moveToRoom(direction component.ExitDirection) {
	if gp.IsMovingToRoom {
		return
	}

	room, exists := gp.MapManager.GetRoomInDirection(string(direction))
	if !exists {
		return
	}

	gp.RoomTarget = rl.NewVector2(
		float32(room.TileX)*service.TileSize+service.TileSize*2,
		float32(room.TileY)*service.TileSize+service.TileSize*2,
	)
	switch direction {
	case component.DirNorth:
		gp.MainChar.Direction = component.DirBack
	case component.DirSouth:
		gp.MainChar.Direction = component.DirFront
	case component.DirEast:
		gp.MainChar.Direction = component.DirRight
	case component.DirWest:
		gp.MainChar.Direction = component.DirLeft
	}
	gp.IsMovingToRoom = true
	gp.RoomViewMgr.IsMoving = true
	gp.MainChar.SetAction(component.ActionWalking)
	gp.MainChar.FrameIndex = 0
	gp.MainChar.FrameTimer = 0
}

func (gp *GamePage) syncRoomContext() {
	if gp.MapManager == nil || gp.MapManager.ActiveRoom == nil {
		return
	}

	room := gp.MapManager.ActiveRoom
	exits := make(map[component.ExitDirection]bool, len(room.Exits))
	for direction := range room.Exits {
		exits[component.ExitDirection(strings.ToUpper(direction))] = true
	}
	gp.RoomViewMgr.CurrentRoom = component.RoomViewData{
		Title:       room.Name,
		Description: room.Description,
		Exits:       exits,
	}

	items := make([]service.RoomItemDetail, 0, len(room.Items))
	for _, itemID := range room.Items {
		items = append(items, service.RoomItemDetail{
			ID:   itemID,
			Name: roomEntityName(itemID),
		})
	}

	npcs := make([]service.RoomNPCDetail, 0, len(room.NPCs))
	for _, npcKey := range room.NPCs {
		kind := "person"
		switch strings.ToLower(npcKey) {
		case "guard":
			kind = "guard"
		case "seller", "shop":
			kind = "seller"
		}
		npcs = append(npcs, service.RoomNPCDetail{
			Key:  npcKey,
			Name: roomEntityName(npcKey),
			Kind: kind,
		})
	}
	gp.RoomDetailsMgr.SetRoomDetails(room.Name, items, npcs)
}

func roomEntityName(value string) string {
	if separator := strings.LastIndex(value, ":"); separator >= 0 {
		value = value[separator+1:]
	}
	value = strings.NewReplacer("_", " ", "-", " ").Replace(value)
	if value == "" {
		return value
	}
	return strings.ToUpper(value[:1]) + value[1:]
}

func (gp *GamePage) Render() {
	gp.EventListener()
	dt := rl.GetFrameTime()
	gp.Update(dt)
	modalOpen := gp.hasOpenModal()
	component.SetButtonInputEnabled(!modalOpen)

	rl.ClearBackground(utils.ThemeBackground)

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
	rl.DrawRectangleLinesEx(rl.NewRectangle(0, 0, topMapSize, topMapSize), 2, utils.ThemeBorder)

	if gp.RoomDetailsComp != nil {
		roomDetailsRect := rl.NewRectangle(topMapSize+6, 8, topDetailsW-16, topMapSize-16)
		gp.RoomDetailsComp.Render(roomDetailsRect)
	}

	rl.DrawLineEx(rl.NewVector2(0, topMapSize), rl.NewVector2(totalW, topMapSize), 2, utils.ThemeBorder)

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

	component.SetButtonInputEnabled(true)

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
