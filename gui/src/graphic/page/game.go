package page

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
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

// GamePage coordinates gameplay state, server commands, and game components.
type GamePage struct {
	BasePage
	MapManager       *service.MapManager
	InventoryMgr     *service.InventoryManager
	ChatMgr          *service.ChatManager
	QuestMgr         *service.QuestManager
	DialogMgr        *service.ActionManager
	RoomDetailsMgr   *service.RoomDetailsManager
	RoomItemMgr      *service.RoomItemManager
	RoomNPCMgr       *service.RoomNPCManager
	AttackMgr        *service.AttackManager
	RoomViewMgr      *component.RoomViewManager
	PlayerMgr        *service.PlayerManager
	MapComp          *component.MapComponent
	MainChar         *character.MainCharacter
	HPBar            *ui.HPBar
	InventoryComp    *component.InventoryComponent
	ChatComp         *component.ChatComponent
	QuestComp        *component.QuestComponent
	DialogComp       *component.ActionComponent
	RoomDetailsComp  *component.RoomDetailsComponent
	AttackComp       *component.AttackComponent
	RoomViewComp     *component.RoomViewComponent
	PlayerComp       *component.PlayerComponent
	ActionBtnsComp   *component.ActionButtonsComponent
	LogViewComp      *component.LogViewComponent
	PathResolver     *utils.PathResolver
	WindowWidth      int32
	WindowHeight     int32
	Camera           rl.Camera2D
	IsMovingToRoom   bool
	movementStart    rl.Vector2
	movementTarget   rl.Vector2
	movementElapsed  float32
	movementDuration float32
	pendingLook      *lookResponse
	commandQueue     chan string
	commandResults   chan commandResult
	commandContext   context.Context
	cancelCommands   context.CancelFunc
}

type commandResult struct {
	command  string
	response string
	err      error
}

type lookResponse struct {
	Room struct {
		ID          string            `json:"id"`
		Name        string            `json:"name"`
		Description string            `json:"description"`
		Exits       map[string]string `json:"exits"`
	} `json:"room"`
	Players []string `json:"players"`
	Items   []string `json:"items"`
	NPCs    []string `json:"npcs"`
}

type serverQuest struct {
	QuestID  string `json:"quest_id"`
	Status   string `json:"status"`
	Progress string `json:"progress"`
}

// NewGamePage creates the game screen using the supplied window size and shared context.
// winWidth and winHeight are the window dimensions in pixels; ctx holds shared game state.
func NewGamePage(winWidth, winHeight int32, ctx *model.GameContext) *GamePage {
	var resourceManager *utils.ResourceManager
	if ctx != nil {
		resourceManager = ctx.ResourceManager
	}
	pathResolver := utils.NewPathResolverWithResourceManager("assets", resourceManager)
	mapJSON := ""
	if ctx != nil {
		mapJSON = ctx.Config.MapJSON
	}
	mapManager := service.NewMapManager(pathResolver, mapJSON, npc.EntityFactory{
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
	roomItemMgr := service.NewRoomItemManager()
	roomNPCMgr := service.NewRoomNPCManager()
	roomDetailsMgr := service.NewRoomDetailsManager(roomItemMgr, roomNPCMgr)
	attackMgr := service.NewAttackManager()
	roomViewMgr := component.NewRoomViewManager()
	playerMgr := service.NewPlayerManager()

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
		BasePage:        BasePage{Context: ctx},
		WindowWidth:     winWidth,
		WindowHeight:    winHeight,
		PathResolver:    pathResolver,
		MapManager:      mapManager,
		InventoryMgr:    inventoryMgr,
		ChatMgr:         chatMgr,
		QuestMgr:        questMgr,
		DialogMgr:       dialogMgr,
		RoomDetailsMgr:  roomDetailsMgr,
		RoomItemMgr:     roomItemMgr,
		RoomNPCMgr:      roomNPCMgr,
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
		commandQueue:   make(chan string, 32),
		commandResults: make(chan commandResult, 32),
	}
	gp.commandContext, gp.cancelCommands = context.WithCancel(context.Background())
	gp.State = pagemodel.GamePage
	gp.RoomViewMgr.OnMove = gp.moveToRoom
	gp.ActionBtnsComp.OnLog = gp.addLog
	gp.ActionBtnsComp.OnCommand = gp.submitCommand
	gp.InventoryComp.OnLog = gp.addLog
	gp.QuestComp.OnLog = gp.addLog
	gp.InventoryComp.OnCommand = gp.submitCommand
	gp.RoomDetailsComp.OnLog = gp.addLog
	gp.RoomDetailsComp.OnTakeItem = func(roomItem service.RoomItemDetail) {
		gp.submitCommand("TAKE " + roomItem.ID)
	}
	gp.DialogComp.OnLog = gp.addLog
	gp.DialogComp.OnAction = func(npcConfig npc.NPCConfig, action string) {
		switch action {
		case "TALK":
			gp.submitCommand("TALK " + npcConfig.Key)
		case "QUEST":
			gp.submitCommand("QUEST " + npcConfig.Key)
		}
	}
	gp.ChatComp.OnLog = gp.addLog
	gp.ChatComp.OnCommand = gp.submitCommand
	gp.AttackComp.OnCommand = func(command string) {
		if command == "FLEE" {
			for _, direction := range []string{"north", "south", "east", "west"} {
				if _, exists := gp.MapManager.ActiveRoom.Exits[direction]; exists {
					command += " " + direction
					break
				}
			}
		}
		gp.submitCommand(command)
	}
	gp.RoomDetailsMgr.OnOpenActions = func(npc npc.NPCConfig) {
		gp.DialogMgr.OpenAction(npc)
		gp.submitCommand("TALK " + npc.Key)
	}
	gp.DialogMgr.HasQuest = func(npcConfig npc.NPCConfig) bool {
		return gp.QuestMgr.HasQuest(npcConfig.QuestID)
	}
	gp.RoomDetailsMgr.OnAttackNPC = func(npc npc.NPCConfig) {
		gp.AttackMgr.StartCombat(npc)
	}
	go gp.commandWorker()
	gp.submitCommand("LOOK")
	gp.submitCommand("STATUS")
	gp.submitCommand("INVENTORY")
	gp.submitCommand("QUESTS")

	return gp
}

func (gp *GamePage) commandWorker() {
	for {
		select {
		case <-gp.commandContext.Done():
			return
		case command := <-gp.commandQueue:
			response, err := gp.Context.TCPClient.Command(gp.commandContext, command)
			select {
			case gp.commandResults <- commandResult{command: command, response: response, err: err}:
			case <-gp.commandContext.Done():
				return
			}
		}
	}
}

func (gp *GamePage) submitCommand(command string) {
	command = strings.TrimSpace(command)
	if command == "" {
		return
	}
	if gp.Context == nil || gp.Context.TCPClient == nil || !gp.Context.TCPClient.IsConnected() {
		gp.addLog("Command not sent: server connection is unavailable.")
		return
	}
	select {
	case gp.commandQueue <- command:
		gp.addLog("You: " + command)
	default:
		gp.addLog("Command not sent: the server request queue is full.")
	}
}

func (gp *GamePage) handleCommandResult(result commandResult) {
	if result.err != nil {
		gp.addLog(fmt.Sprintf("%s failed: %v", result.command, result.err))
		if commandName(result.command) == "MOVE" {
			gp.IsMovingToRoom = false
			gp.RoomViewMgr.IsMoving = false
			gp.MainChar.SetAction(character.ActionIdle)
		}
		return
	}
	gp.addLog("Server: " + result.response)
	if strings.HasPrefix(result.response, "ERR ") {
		if commandName(result.command) == "MOVE" {
			gp.IsMovingToRoom = false
			gp.RoomViewMgr.IsMoving = false
			gp.MainChar.SetAction(character.ActionIdle)
		}
		return
	}

	body := strings.TrimPrefix(result.response, "OK ")
	fields := strings.Fields(result.command)
	if len(fields) == 0 {
		return
	}
	switch strings.ToUpper(fields[0]) {
	case "LOOK":
		var data lookResponse
		if err := json.Unmarshal([]byte(body), &data); err != nil {
			gp.addLog("Could not decode LOOK response: " + err.Error())
			return
		}
		if gp.IsMovingToRoom {
			gp.pendingLook = &data
			return
		}
		gp.applyLook(data)
	case "MOVE":
		roomID := strings.TrimPrefix(body, "room=")
		if !gp.beginRoomTransition(roomID) {
			gp.addLog("Server moved to an unknown room: " + roomID)
			gp.IsMovingToRoom = false
			gp.RoomViewMgr.IsMoving = false
			return
		}
		gp.submitCommand("LOOK")
	case "INVENTORY":
		var itemIDs []string
		if err := json.Unmarshal([]byte(body), &itemIDs); err != nil {
			gp.addLog("Could not decode INVENTORY response: " + err.Error())
			return
		}
		gp.applyInventory(itemIDs)
	case "QUESTS":
		var serverQuests []serverQuest
		if err := json.Unmarshal([]byte(body), &serverQuests); err != nil {
			gp.addLog("Could not decode QUESTS response: " + err.Error())
			return
		}
		gp.applyQuests(serverQuests)
	case "STATUS":
		var status struct {
			HP     int    `json:"hp"`
			MaxHP  int    `json:"max_hp"`
			Status string `json:"status"`
		}
		if err := json.Unmarshal([]byte(body), &status); err != nil {
			gp.addLog("Could not decode STATUS response: " + err.Error())
			return
		}
		gp.applyStatus(status.HP, status.MaxHP)
	case "TAKE", "DROP":
		gp.submitCommand("LOOK")
		gp.submitCommand("INVENTORY")
	case "QUEST":
		var questResult struct {
			Status string `json:"status"`
		}
		if err := json.Unmarshal([]byte(body), &questResult); err == nil && questResult.Status != "" {
			gp.DialogMgr.CurrentText = "Quest status: " + questResult.Status
		} else {
			gp.DialogMgr.CurrentText = body
		}
		gp.DialogMgr.SetQuestAccepted(true)
		gp.submitCommand("QUESTS")
		gp.submitCommand("INVENTORY")
	case "ATTACK", "DEFEND", "FLEE":
		status := gp.applyCombatResult(strings.ToUpper(fields[0]), body)
		gp.submitCommand("STATUS")
		if strings.ToUpper(fields[0]) == "FLEE" || status == "respawned" {
			gp.AttackMgr.IsOpen = false
			gp.submitCommand("LOOK")
		}
	case "TALK":
		if gp.DialogMgr.IsOpen {
			gp.DialogMgr.CurrentText = body
		}
	case "QUIT":
		gp.returnToMainMenu()
	}
}

func (gp *GamePage) applyLook(data lookResponse) {
	if !gp.MapManager.SetActiveRoom(data.Room.ID) {
		gp.addLog("LOOK returned a room missing from the server map: " + data.Room.ID)
		return
	}
	room := gp.MapManager.ActiveRoom
	room.Name = data.Room.Name
	room.Description = data.Room.Description
	room.Exits = data.Room.Exits
	gp.RoomViewMgr.CurrentRoom = component.RoomViewData{
		Title: data.Room.Name, Description: data.Room.Description,
		Exits: make(map[component.ExitDirection]bool, len(data.Room.Exits)),
	}
	for direction := range data.Room.Exits {
		gp.RoomViewMgr.CurrentRoom.Exits[component.ExitDirection(strings.ToUpper(direction))] = true
	}

	items := make([]service.RoomItemDetail, 0, len(data.Items))
	for _, itemID := range data.Items {
		item, exists := gp.MapManager.GetItemConfig(itemID)
		name, description := roomEntityName(itemID), ""
		if exists {
			name, description = item.Name, item.Description
		}
		items = append(items, service.RoomItemDetail{ID: itemID, Name: name, Description: description})
	}
	gp.RoomItemMgr.SetRoomItems(room.ID, items)

	npcs := make([]npc.NPCConfig, 0, len(data.NPCs))
	for _, key := range data.NPCs {
		config, exists := gp.MapManager.GetNPCConfig(key)
		if !exists {
			config = npc.NPCConfig{Key: key, Name: roomEntityName(key), ImageType: key}
		}
		npcs = append(npcs, config)
	}
	gp.RoomNPCMgr.SetRoomNPCs(room.ID, npcs)
	gp.RoomDetailsMgr.SetActiveRoom(room.ID, data.Room.Name)
	gp.PlayerMgr.SetPlayers(data.Players)
	if !gp.IsMovingToRoom {
		gp.setPlayerRoom(room.ID)
	}
}

func (gp *GamePage) setPlayerRoom(roomID string) {
	if !gp.MapManager.SetActiveRoom(roomID) {
		return
	}
	room := gp.MapManager.ActiveRoom
	position := rl.NewVector2(float32(room.TileX)*service.TileSize+service.TileSize*2, float32(room.TileY)*service.TileSize+service.TileSize*2)
	gp.MainChar.SetPosition(position.X, position.Y)
	gp.Camera.Target = position
	gp.RoomViewMgr.IsMoving = false
	gp.IsMovingToRoom = false
}

func (gp *GamePage) beginRoomTransition(roomID string) bool {
	if !gp.MapManager.SetActiveRoom(roomID) {
		return false
	}

	room := gp.MapManager.ActiveRoom
	gp.movementStart = gp.MainChar.Position
	gp.movementTarget = rl.NewVector2(
		float32(room.TileX)*service.TileSize+service.TileSize*2,
		float32(room.TileY)*service.TileSize+service.TileSize*2,
	)
	dx := gp.movementTarget.X - gp.movementStart.X
	dy := gp.movementTarget.Y - gp.movementStart.Y
	distance := float32(math.Hypot(float64(dx), float64(dy)))
	if gp.MainChar.Speed <= 0 {
		return false
	}

	gp.movementElapsed = 0
	gp.movementDuration = distance / gp.MainChar.Speed
	gp.pendingLook = nil
	gp.setFacingDirection(dx, dy)
	gp.IsMovingToRoom = distance > 0
	gp.RoomViewMgr.IsMoving = gp.IsMovingToRoom
	if !gp.IsMovingToRoom {
		gp.MainChar.SetPosition(gp.movementTarget.X, gp.movementTarget.Y)
	}
	return true
}

func (gp *GamePage) setFacingDirection(dx, dy float32) {
	if math.Abs(float64(dx)) > math.Abs(float64(dy)) {
		if dx > 0 {
			gp.MainChar.Direction = character.DirRight
		} else {
			gp.MainChar.Direction = character.DirLeft
		}
		return
	}
	if dy > 0 {
		gp.MainChar.Direction = character.DirFront
	} else {
		gp.MainChar.Direction = character.DirBack
	}
}

func (gp *GamePage) updateRoomTransition(dt float32) {
	gp.movementElapsed += dt
	progress := gp.movementElapsed / gp.movementDuration
	if progress >= 1 {
		gp.MainChar.SetPosition(gp.movementTarget.X, gp.movementTarget.Y)
		gp.IsMovingToRoom = false
		gp.RoomViewMgr.IsMoving = false
		gp.MainChar.SetAction(character.ActionIdle)
		if gp.pendingLook != nil {
			gp.applyLook(*gp.pendingLook)
			gp.pendingLook = nil
		}
		return
	}

	gp.MainChar.Position = rl.NewVector2(
		gp.movementStart.X+(gp.movementTarget.X-gp.movementStart.X)*progress,
		gp.movementStart.Y+(gp.movementTarget.Y-gp.movementStart.Y)*progress,
	)
	gp.MainChar.SetAction(character.ActionWalking)
	gp.MainChar.AdvanceFrame(dt, true)
}

func commandName(command string) string {
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return ""
	}
	return strings.ToUpper(fields[0])
}

func (gp *GamePage) applyInventory(itemIDs []string) {
	items := make([]inventory.InventoryItem, 0, len(itemIDs))
	indexes := make(map[string]int, len(itemIDs))
	for _, itemID := range itemIDs {
		if index, exists := indexes[itemID]; exists {
			items[index].Count++
			continue
		}
		config, exists := gp.MapManager.GetItemConfig(itemID)
		name, cost := roomEntityName(itemID), 0
		if exists {
			name, cost = config.Name, config.Cost
		}
		indexes[itemID] = len(items)
		items = append(items, inventory.InventoryItem{
			ID: itemID, Name: name, Type: inventory.InventoryItemType(itemID),
			Category: inventory.InventoryCategory(config.Category), Cost: cost, Count: 1,
		})
	}
	gp.InventoryMgr.SetItems(items)
}

func (gp *GamePage) applyQuests(serverQuests []serverQuest) {
	quests := make([]*service.Quest, 0, len(serverQuests))
	for _, serverQuest := range serverQuests {
		config, exists := gp.MapManager.GetQuestConfig(serverQuest.QuestID)
		title, giver := serverQuest.QuestID, ""
		if exists {
			title = config.Title
			if npcConfig, found := gp.MapManager.GetNPCConfig(config.Giver); found {
				giver = npcConfig.Name
			} else {
				giver = config.Giver
			}
		}
		current, target := parseProgress(serverQuest.Progress)
		quests = append(quests, &service.Quest{
			ID: serverQuest.QuestID, Title: title, NPCName: giver,
			CurrentProgress: current, TargetProgress: target,
		})
	}
	gp.QuestMgr.SetQuests(quests)
}

func parseProgress(progress string) (int, int) {
	current, target, found := strings.Cut(progress, "/")
	if !found {
		return 0, 1
	}
	currentValue, currentErr := strconv.Atoi(current)
	targetValue, targetErr := strconv.Atoi(target)
	if currentErr != nil || targetErr != nil || targetValue < 1 {
		return 0, 1
	}
	return currentValue, targetValue
}

func (gp *GamePage) applyStatus(hp, maxHP int) {
	if maxHP < 1 {
		return
	}
	gp.AttackMgr.PlayerHP = hp
	gp.AttackMgr.PlayerMaxHP = maxHP
	gp.HPBar.SetMaxHP(float32(maxHP))
	gp.HPBar.SetHP(float32(hp))
}

func (gp *GamePage) applyCombatResult(command, body string) string {
	var result struct {
		HP       int    `json:"hp"`
		Attacker int    `json:"attacker_hp"`
		TargetHP int    `json:"target_hp"`
		Damage   int    `json:"damage"`
		Status   string `json:"status"`
	}
	if err := json.Unmarshal([]byte(body), &result); err != nil {
		return ""
	}
	playerHP := result.HP
	if result.Attacker > 0 {
		playerHP = result.Attacker
	}
	if playerHP > 0 {
		gp.AttackMgr.PlayerHP = playerHP
		gp.HPBar.SetHP(float32(playerHP))
	}
	if command == "ATTACK" && gp.AttackMgr.NPC.Key != "" {
		gp.AttackMgr.NPCHP = result.TargetHP
		gp.RoomNPCMgr.DamageActiveNPC(gp.AttackMgr.NPC.Key, result.Damage)
	}
	if result.Status == "victory" || result.Status == "respawned" {
		gp.AttackMgr.IsCombatOver = true
		gp.AttackMgr.IsOpen = false
	}
	gp.AttackMgr.AddLog(body)
	return result.Status
}

func (gp *GamePage) handleServerEvent(event string) {
	parts := strings.SplitN(event, " ", 5)
	if gp.ChatMgr != nil && len(parts) == 5 && parts[0] == "EVT" && parts[2] == "CHAT" {
		channel := service.ChannelGlobal
		switch parts[1] {
		case "ROOM":
			channel = service.ChannelRoom
		case "GROUP":
			channel = service.ChannelGroup
		}
		gp.ChatMgr.AddMessage(channel, parts[3], parts[4], parts[3] == gp.Context.PlayerName)
	}
	if len(parts) == 5 && parts[0] == "EVT" && parts[1] == "ROOM" && parts[2] == "PRESENCE" {
		switch parts[3] {
		case "ENTER":
			for _, player := range gp.PlayerMgr.Players {
				if player.Name == parts[4] {
					gp.addLog(event)
					return
				}
			}
			gp.PlayerMgr.AddPlayer(parts[4], parts[4])
		case "LEAVE":
			for _, player := range gp.PlayerMgr.Players {
				if player.Name == parts[4] {
					gp.PlayerMgr.RemovePlayer(player.ID)
					break
				}
			}
		}
	}
	gp.addLog(event)
}

// Unload releases resources owned by the game page.
func (gp *GamePage) Unload() {
	if gp.IsUnloaded {
		return
	}
	if gp.cancelCommands != nil {
		gp.cancelCommands()
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

// returnToMainMenu closes the server connection and requests the main menu.
func (gp *GamePage) returnToMainMenu() {
	if gp.Context != nil && gp.Context.TCPClient != nil {
		if err := gp.Context.TCPClient.Close(); err != nil {
			gp.addLog("Could not disconnect from the server.")
		} else {
			gp.addLog("Disconnected from the server.")
		}
	}
	gp.addLog("Returned to the main menu.")
	gp.NextState = pagemodel.MainMenu
}

// hasOpenModal reports whether any gameplay modal is currently open.
func (gp *GamePage) hasOpenModal() bool {
	return (gp.QuestMgr != nil && gp.QuestMgr.IsOpen) ||
		(gp.InventoryMgr != nil && gp.InventoryMgr.IsOpen) ||
		(gp.ChatMgr != nil && gp.ChatMgr.IsOpen) ||
		(gp.DialogMgr != nil && gp.DialogMgr.IsOpen) ||
		(gp.AttackMgr != nil && gp.AttackMgr.IsOpen)
}

// Update advances game state by dt seconds.
func (gp *GamePage) Update(dt float32) {
	if gp.Context != nil && gp.Context.TCPClient != nil {
		for {
			select {
			case result := <-gp.commandResults:
				gp.handleCommandResult(result)
			case event := <-gp.Context.TCPClient.Events():
				gp.handleServerEvent(event)
			default:
				goto drained
			}
		}
	}
drained:
	if gp.IsMovingToRoom {
		gp.updateRoomTransition(dt)
	} else if gp.MainChar.IsDead || gp.hasOpenModal() {
		gp.MainChar.SetAction(character.ActionIdle)
		gp.MainChar.AdvanceFrame(dt, false)
	} else if gp.MainChar.Action == character.ActionAttacking || gp.MainChar.Action == character.ActionHurt {
		if gp.MainChar.AdvanceFrame(dt, false) {
			gp.MainChar.SetAction(character.ActionIdle)
		}
	} else {
		gp.MainChar.SetAction(character.ActionIdle)
		gp.MapManager.Update(dt, gp.MainChar.Position)
		gp.MainChar.AdvanceFrame(dt, true)
	}
	gp.Camera.Target = gp.MainChar.Position
	if gp.HPBar != nil {
		gp.HPBar.Update(dt)
	}
}

// moveToRoom starts player movement toward the adjacent room in direction.
func (gp *GamePage) moveToRoom(direction component.ExitDirection) {
	if gp.IsMovingToRoom {
		return
	}

	_, exists := gp.MapManager.GetRoomInDirection(string(direction))
	if !exists {
		gp.addLog(fmt.Sprintf("There is no room to the %s.", strings.ToLower(string(direction))))
		return
	}

	gp.RoomViewMgr.IsMoving = true
	gp.IsMovingToRoom = true
	gp.submitCommand("MOVE " + strings.ToLower(string(direction)))
}

// addLog appends message to the game's log when the log component is available.
func (gp *GamePage) addLog(message string) {
	if gp.LogViewComp != nil {
		gp.LogViewComp.AddLog(message)
	}
}

// roomEntityName formats an entity identifier as a display name.
// value is the identifier to normalize.
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

// Render draws the game page and its components.
func (gp *GamePage) Render() {
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
