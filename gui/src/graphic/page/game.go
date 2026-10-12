package page

import (
	"fmt"
	"strings"

	"tap-gui/src/graphic/component"
	"tap-gui/src/graphic/component/character"
	"tap-gui/src/graphic/component/ui"
	"tap-gui/src/model"
	"tap-gui/src/model/inventory"
	"tap-gui/src/model/npc"
	pagemodel "tap-gui/src/model/page"
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
	DialogMgr       *service.ActionManager
	RoomDetailsMgr  *service.RoomDetailsManager
	AttackMgr       *service.AttackManager
	RoomViewMgr     *component.RoomViewManager
	PlayerMgr       *component.PlayerManager
	MapComp         *component.MapComponent
	MainChar        *character.MainCharacter
	HPBar           *ui.HPBar
	InventoryComp   *component.InventoryComponent
	ChatComp        *component.ChatComponent
	QuestComp       *component.QuestComponent
	DialogComp      *component.ActionComponent
	RoomDetailsComp *component.RoomDetailsComponent
	AttackComp      *component.AttackComponent
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
	mapManager := service.NewMapManager(pathResolver, "", npc.EntityFactory{
		NewMonster: func(scale float32, resolver *utils.PathResolver, assetPath, name string) npc.Monster {
			return character.NewMonster(scale, resolver, assetPath, name)
		},
		NewNPC: func(kind, relativeDir string, scale float32) (npc.INpc, error) {
			switch kind {
			case "guard":
				return character.NewGuardCharacter(pathResolver, relativeDir, scale)
			case "merchant":
				return character.NewMerchantCharacter(pathResolver, relativeDir, scale)
			default:
				return character.NewPersonCharacter(pathResolver, relativeDir, scale)
			}
		},
	})

	questMgr := service.NewQuestManager()
	inventoryMgr := service.NewInventoryManager()
	chatMgr := service.NewChatManager()
	dialogMgr := service.NewActionManager()
	roomDetailsMgr := service.NewRoomDetailsManager()
	attackMgr := service.NewAttackManager()
	roomViewMgr := component.NewRoomViewManager()
	playerMgr := component.NewPlayerManager()

	mapComp := component.NewMapComponent(mapManager)
	inventoryComp := component.NewInventoryComponent(inventoryMgr, pathResolver)
	chatComp := component.NewChatComponent(chatMgr)
	questComp := component.NewQuestComponent(questMgr)
	dialogComp := component.NewActionComponent(dialogMgr, pathResolver)
	roomDetailsComp := component.NewRoomDetailsComponent(roomDetailsMgr, pathResolver)
	roomViewComp := component.NewRoomViewComponent(roomViewMgr)
	playerComp := component.NewPlayerComponent(playerMgr)
	actionBtnsComp := component.NewActionButtonsComponent(inventoryMgr, chatMgr, questMgr)
	logViewComp := component.NewLogViewComponent()

	startX, startY := mapManager.GetStartPosition()

	mainChar := character.NewMainCharacter(service.CharScale, pathResolver, pathResolver.Resolve("characters", "main_character"))
	mainChar.SetPosition(startX, startY)
	hpBar := ui.NewHPBar(30, 15, 180, 20)

	attackComp := component.NewAttackComponent(attackMgr, pathResolver)

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
		DialogMgr:       dialogMgr,
		RoomDetailsMgr:  roomDetailsMgr,
		AttackMgr:       attackMgr,
		RoomViewMgr:     roomViewMgr,
		PlayerMgr:       playerMgr,
		MapComp:         mapComp,
		MainChar:        mainChar,
		HPBar:           hpBar,
		InventoryComp:   inventoryComp,
		ChatComp:        chatComp,
		QuestComp:       questComp,
		DialogComp:      dialogComp,
		RoomDetailsComp: roomDetailsComp,
		AttackComp:      attackComp,
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
	gp.State = pagemodel.GamePage
	gp.RoomViewMgr.OnMove = gp.moveToRoom

	gp.RoomDetailsMgr.OnOpenActions = func(npc npc.NPCConfig) {
		gp.DialogMgr.OpenAction(npc)
	}
	gp.RoomDetailsMgr.OnAttackNPC = func(npc npc.NPCConfig) {
		gp.AttackMgr.StartCombat(npc)
	}

	gp.AttackMgr.OnMonsterDefeated = func(npc npc.NPCConfig) {
		updatedNPC, attacked := gp.MapManager.DamageNPC(npc.Key, 999)
		if attacked {
			for index := range gp.RoomDetailsMgr.NPCs {
				if gp.RoomDetailsMgr.NPCs[index].Key == npc.Key {
					gp.RoomDetailsMgr.NPCs[index] = updatedNPC
					break
				}
			}
		}
		gp.MainChar.SetPosition(gp.MainChar.Position.X, gp.MainChar.Position.Y)
	}

	gp.AttackMgr.OnPlayerRespawn = func() {
		defX, defY := gp.MapManager.GetStartPosition()
		gp.MainChar.IsDead = false
		gp.MainChar.SetAction(character.ActionIdle)
		gp.MainChar.SetPosition(defX, defY)
		gp.Camera.Target = gp.MainChar.Position
		gp.IsMovingToRoom = false
		gp.RoomViewMgr.IsMoving = false
		gp.HPBar.SetHP(50)
	}

	gp.AttackMgr.OnFlee = func() {
		defX, defY := gp.MapManager.GetStartPosition()
		gp.MainChar.SetPosition(defX, defY)
	}

	gp.AttackMgr.OnOpenInventory = func() {
		if gp.InventoryMgr != nil {
			gp.InventoryMgr.InCombatMode = true
			gp.InventoryMgr.IsOpen = true
		}
	}

	gp.InventoryMgr.OnUseItemInCombat = func(item inventory.InventoryItem) {
		gp.AttackMgr.PlayerHP += item.Cost
		if gp.AttackMgr.PlayerHP > gp.AttackMgr.PlayerMaxHP {
			gp.AttackMgr.PlayerHP = gp.AttackMgr.PlayerMaxHP
		}
		gp.AttackMgr.AddLog(fmt.Sprintf("Used %s! Recovered %d HP.", item.Name, item.Cost))
	}

	gp.DialogMgr.HasQuest = func(npc npc.NPCConfig) bool {
		return gp.QuestMgr.HasQuest(npc.QuestID)
	}
	gp.DialogMgr.OnAcceptQuest = func(npc npc.NPCConfig) bool {
		return gp.QuestMgr.AcceptQuest(npc)
	}

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
	if gp.DialogComp != nil {
		gp.DialogComp.Unload()
	}
	if gp.AttackComp != nil {
		gp.AttackComp.Unload()
	}
	gp.BasePage.Unload()
}

func (gp *GamePage) EventListener() {
	if rl.IsKeyPressed(rl.KeyEscape) && !gp.hasOpenModal() {
		gp.NextState = pagemodel.MainMenu
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
		(gp.ChatMgr != nil && gp.ChatMgr.IsOpen) ||
		(gp.DialogMgr != nil && gp.DialogMgr.IsOpen) ||
		(gp.AttackMgr != nil && gp.AttackMgr.IsOpen)
}

func (gp *GamePage) Update(dt float32) {
	if gp.AttackMgr != nil {
		gp.AttackMgr.Update(dt)
	}

	if gp.MainChar.IsDead {
		gp.MainChar.AdvanceFrame(dt, false)
		return
	}

	if gp.AttackMgr != nil && gp.AttackMgr.IsOpen {
		gp.HPBar.SetHP(float32(gp.AttackMgr.PlayerHP))
	}

	if gp.hasOpenModal() {
		gp.MainChar.SetAction(character.ActionIdle)
		return
	}

	if gp.MainChar.Action == character.ActionAttacking || gp.MainChar.Action == character.ActionHurt {
		if gp.MainChar.AdvanceFrame(dt, false) {
			gp.MainChar.SetAction(character.ActionIdle)
		}
		return
	}

	if rl.IsKeyPressed(rl.KeySpace) {
		gp.MainChar.SetAction(character.ActionAttacking)
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
			gp.MainChar.SetAction(character.ActionIdle)
		} else {
			moveX := delta.X / distance
			moveY := delta.Y / distance
			gp.MainChar.Position.X += moveX * step
			gp.MainChar.Position.Y += moveY * step
			if moveX < 0 {
				gp.MainChar.Direction = character.DirLeft
			} else if moveX > 0 {
				gp.MainChar.Direction = character.DirRight
			} else if moveY < 0 {
				gp.MainChar.Direction = character.DirBack
			} else {
				gp.MainChar.Direction = character.DirFront
			}
			gp.MainChar.SetAction(character.ActionWalking)
		}
	} else {
		gp.MainChar.SetAction(character.ActionIdle)
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
		gp.MainChar.Direction = character.DirBack
	case component.DirSouth:
		gp.MainChar.Direction = character.DirFront
	case component.DirEast:
		gp.MainChar.Direction = character.DirRight
	case component.DirWest:
		gp.MainChar.Direction = character.DirLeft
	}
	gp.IsMovingToRoom = true
	gp.RoomViewMgr.IsMoving = true
	gp.MainChar.SetAction(character.ActionWalking)
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

	npcs := make([]npc.NPCConfig, 0, len(room.NPCs))
	for _, npcKey := range room.NPCs {
		npcCfg, exists := gp.MapManager.GetNPCConfig(npcKey)
		if !exists {
			npcCfg = npc.NPCConfig{
				Key:       npcKey,
				Name:      roomEntityName(npcKey),
				ImageType: npcKey,
			}
		}
		npcs = append(npcs, npcCfg)
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
	ui.SetButtonInputEnabled(!modalOpen)

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

	ui.SetButtonInputEnabled(true)

	if gp.ChatComp != nil {
		gp.ChatComp.Render(gp.WindowWidth, gp.WindowHeight)
	}
	if gp.QuestComp != nil {
		gp.QuestComp.Render(gp.WindowWidth, gp.WindowHeight)
	}
	if gp.DialogComp != nil {
		gp.DialogComp.Render(gp.WindowWidth, gp.WindowHeight)
	}
	if gp.AttackComp != nil {
		gp.AttackComp.Render(gp.WindowWidth, gp.WindowHeight)
	}
	if gp.InventoryComp != nil {
		gp.InventoryComp.Render(gp.WindowWidth, gp.WindowHeight)
	}
}
