package page

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"tap-gui/src/graphic/component"
	"tap-gui/src/model/enums"
	"tap-gui/src/utils"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	BaseTileSize = 64
	GroundScale  = 1.0
	TileSize     = float32(BaseTileSize) * GroundScale

	CharScale   = 0.25
	CellSpacing = 15
)

type RoomConfig struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Exits       map[string]string `json:"exits"`
	Items       []string          `json:"items"`
	NPCs        []string          `json:"npcs"`
}

type MapConfig struct {
	World struct {
		Rooms map[string]RoomConfig `json:"rooms"`
	} `json:"world"`
}

type TileType int

const (
	TileWall TileType = iota
	TilePath
	TileRoom
)

type EnvCategory int

const (
	CategoryScenery EnvCategory = iota
	CategoryBuilding
)

type EnvObjectInfo struct {
	Path     string
	Category EnvCategory
	Scale    float32
	Texture  rl.Texture2D
}

type RoomNode struct {
	ID          string
	Name        string
	Description string
	GridX       int
	GridY       int
	TileX       int
	TileY       int
	Color       rl.Color
	NPCs        []string
	Items       []string
}

type WorldMap struct {
	Width        int
	Height       int
	Grid         [][]TileType
	Rooms        map[string]*RoomNode
	EnvObjectIdx [][]int
}

const defaultMapConfigJSON = `{
    "world": {
        "rooms": {
            "start": {
                "items": ["0:ale"],
                "name": "Village Square",
                "description": "A bustling square with cobblestone paths.",
                "exits": {
                    "north": "tavern",
                    "east": "shop",
                    "south": "forest_edge",
                    "west": "hidden_grove"
                },
                "npcs": ["guard"]
            },
            "tavern": {
                "items": [],
                "name": "Tavern",
                "description": "A lively tavern smelling of ale.",
                "exits": {
                    "south": "start"
                },
                "npcs": ["mayor", "person"]
            },
            "shop": {
                "items": ["156:sword"],
                "name": "Shop",
                "description": "A cluttered shop filled with oddities.",
                "exits": {
                    "west": "start",
                    "south": "deep_forest"
                },
                "npcs": ["seller"]
            },
            "forest_edge": {
                "items": [],
                "name": "Forest Edge",
                "description": "The trees start to thicken here.",
                "exits": {
                    "north": "start",
                    "east": "deep_forest"
                },
                "npcs": ["goblin"]
            },
            "hidden_grove": {
                "items": ["241:herb"],
                "name": "Hidden Grove",
                "description": "A quiet grove with rare plants.",
                "exits": {
                    "east": "start"
                },
                "npcs": []
            },
            "deep_forest": {
                "items": [],
                "name": "Deep Forest",
                "description": "The woods are dark and full of shadows.",
                "exits": {
                    "north": "shop",
                    "west": "forest_edge",
                    "south": "cave_entrance"
                },
                "npcs": ["goblin"]
            },
            "cave_entrance": {
                "items": [],
                "name": "Cave Entrance",
                "description": "A dark hole in the rock face.",
                "exits": {
                    "north": "deep_forest",
                    "south": "boss_lair"
                },
                "npcs": []
            },
            "boss_lair": {
                "items": ["117:magic_amulet"],
                "name": "Boss Lair",
                "description": "A vast cavern filled with bones.",
                "exits": {
                    "north": "cave_entrance"
                },
                "npcs": ["boss"]
            }
        }
    }
}`

func NewWorldMapFromConfig(configJSON string) *WorldMap {
	var config MapConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		_ = json.Unmarshal([]byte(defaultMapConfigJSON), &config)
	}

	roomGridCoords := make(map[string][2]int)
	visited := make(map[string]bool)

	var computeCoords func(roomID string, gx, gy int)
	computeCoords = func(roomID string, gx, gy int) {
		if visited[roomID] {
			return
		}
		visited[roomID] = true
		roomGridCoords[roomID] = [2]int{gx, gy}

		room := config.World.Rooms[roomID]
		for dir, targetID := range room.Exits {
			dx, dy := 0, 0
			switch dir {
			case "north":
				dy = -1
			case "south":
				dy = 1
			case "east":
				dx = 1
			case "west":
				dx = -1
			}
			computeCoords(targetID, gx+dx, gy+dy)
		}
	}

	if _, exists := config.World.Rooms["start"]; exists {
		computeCoords("start", 0, 0)
	}

	minX, maxX, minY, maxY := 0, 0, 0, 0
	for _, coord := range roomGridCoords {
		if coord[0] < minX {
			minX = coord[0]
		}
		if coord[0] > maxX {
			maxX = coord[0]
		}
		if coord[1] < minY {
			minY = coord[1]
		}
		if coord[1] > maxY {
			maxY = coord[1]
		}
	}

	// Enormous padding to create a vast outer landscape around the core map
	padding := 20
	mapWidth := (maxX - minX + 1) * CellSpacing + padding*2
	mapHeight := (maxY - minY + 1) * CellSpacing + padding*2

	grid := make([][]TileType, mapHeight)
	envIdx := make([][]int, mapHeight)

	for y := 0; y < mapHeight; y++ {
		grid[y] = make([]TileType, mapWidth)
		envIdx[y] = make([]int, mapWidth)
		for x := 0; x < mapWidth; x++ {
			grid[y][x] = TileWall // Default outer map (now functioning as Impassable Lawn)

			// 40% chance to render dense forest flora (trees or bushes)
			if (x*17+y*11)%10 < 4 {
				if (x+y)%2 == 0 {
					envIdx[y][x] = 5 // Bushes-Large index
				} else {
					envIdx[y][x] = 6 // Tree-Large index
				}
			} else {
				envIdx[y][x] = -1
			}
		}
	}

	worldMap := &WorldMap{
		Width:        mapWidth,
		Height:       mapHeight,
		Grid:         grid,
		Rooms:        make(map[string]*RoomNode),
		EnvObjectIdx: envIdx,
	}

	for roomID, coord := range roomGridCoords {
		tileX := (coord[0]-minX)*CellSpacing + padding + 4
		tileY := (coord[1]-minY)*CellSpacing + padding + 4

		roomCfg := config.World.Rooms[roomID]
		roomNode := &RoomNode{
			ID:          roomID,
			Name:        roomCfg.Name,
			Description: roomCfg.Description,
			GridX:       coord[0],
			GridY:       coord[1],
			TileX:       tileX,
			TileY:       tileY,
			NPCs:        roomCfg.NPCs,
			Items:       roomCfg.Items,
		}
		worldMap.Rooms[roomID] = roomNode

		for ry := -3; ry <= 3; ry++ {
			for rx := -3; rx <= 3; rx++ {
				if tileY+ry >= 0 && tileY+ry < mapHeight && tileX+rx >= 0 && tileX+rx < mapWidth {
					grid[tileY+ry][tileX+rx] = TileRoom
				}
			}
		}
	}

	for roomID, roomNode := range worldMap.Rooms {
		roomCfg := config.World.Rooms[roomID]
		for _, targetID := range roomCfg.Exits {
			if targetNode, exists := worldMap.Rooms[targetID]; exists {
				carvePath(grid, roomNode.TileX, roomNode.TileY, targetNode.TileX, targetNode.TileY)
			}
		}
	}

	return worldMap
}

func carvePath(grid [][]TileType, x1, y1, x2, y2 int) {
	currX, currY := x1, y1
	for currX != x2 || currY != y2 {
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				if currY+dy >= 0 && currY+dy < len(grid) && currX+dx >= 0 && currX+dx < len(grid[0]) {
					if grid[currY+dy][currX+dx] == TileWall {
						grid[currY+dy][currX+dx] = TilePath
					}
				}
			}
		}
		if currX < x2 {
			currX++
		} else if currX > x2 {
			currX--
		} else if currY < y2 {
			currY++
		} else if currY > y2 {
			currY--
		}
	}
}

type GamePage struct {
	BasePage
	World            *WorldMap
	MainChar         *component.MainCharacter
	Enemies          []*component.Monster
	NPCs             []*component.NPC
	PathTextures     map[string]rl.Texture2D
	BuildingTextures map[string]rl.Texture2D
	EmptyTexture     rl.Texture2D
	SceneryObjects   []EnvObjectInfo
	PathResolver     *utils.PathResolver
	WindowWidth      int32
	WindowHeight     int32
	ActiveRoom       *RoomNode
	Camera           rl.Camera2D
}

func NewGamePage(winWidth, winHeight int32) *GamePage {
	gp := &GamePage{
		WindowWidth:  winWidth,
		WindowHeight: winHeight,
		PathResolver: utils.NewPathResolver("assets"),
		PathTextures: make(map[string]rl.Texture2D),
	}
	gp.State = enums.GamePage

	gp.loadGroundTextures()
	gp.loadBuildingTextures()
	gp.loadEnvironmentObjects()

	gp.World = NewWorldMapFromConfig(defaultMapConfigJSON)

	startRoom, exists := gp.World.Rooms["start"]
	startX := float32(1) * TileSize
	startY := float32(1) * TileSize
	if exists {
		startX = float32(startRoom.TileX)*TileSize + TileSize/2
		startY = float32(startRoom.TileY)*TileSize + TileSize/2
		gp.ActiveRoom = startRoom
	}

	gp.MainChar = component.NewMainCharacter(startX, startY, CharScale, gp.PathResolver.Resolve("characters", "main_character"))

	gp.Enemies = make([]*component.Monster, 0)
	gp.NPCs = make([]*component.NPC, 0)

	for _, room := range gp.World.Rooms {
		roomX := float32(room.TileX)*TileSize + TileSize/2
		roomY := float32(room.TileY)*TileSize + TileSize/2

		for _, npcName := range room.NPCs {
			if strings.Contains(strings.ToLower(npcName), "boss") {
				e := component.NewMonster(roomX+100, roomY+100, CharScale, gp.PathResolver.Resolve("characters", "boss"), "Boss")
				gp.Enemies = append(gp.Enemies, e)
			} else if strings.Contains(strings.ToLower(npcName), "goblin") {
				e := component.NewMonster(roomX+100, roomY-100, CharScale, gp.PathResolver.Resolve("characters", "gobelin"), "Goblin")
				gp.Enemies = append(gp.Enemies, e)
			} else {
				// Initialize component.NPC with correct dynamically resolved texture paths
				npcPath := gp.PathResolver.Resolve("characters", strings.ToLower(npcName))
				n := component.NewNPC(roomX-120, roomY+80, CharScale, npcPath, npcName)
				gp.NPCs = append(gp.NPCs, n)
			}
		}
	}

	gp.Camera = rl.Camera2D{
		Target:   gp.MainChar.Position,
		Offset:   rl.NewVector2(float32(winWidth)/2, float32(winHeight)/2),
		Rotation: 0.0,
		Zoom:     1.0,
	}

	return gp
}

func (gp *GamePage) loadGroundTextures() {
	bitmasks := []string{
		"0000", "0001", "0010", "0011",
		"0100", "0101", "0110", "0111",
		"1000", "1001", "1010", "1011",
		"1100", "1101", "1110", "1111",
	}
	for _, mask := range bitmasks {
		path := gp.PathResolver.Resolve("ground", mask+".png")
		gp.PathTextures[mask] = rl.LoadTexture(path)
	}
	gp.EmptyTexture = rl.LoadTexture(gp.PathResolver.Resolve("ground", "lawn.png"))
}

func (gp *GamePage) loadBuildingTextures() {
	gp.BuildingTextures = make(map[string]rl.Texture2D)
	buildings := []string{"House.png", "shop.png", "Tavern.png", "Castle-Round.png", "Tent.png"}
	for _, b := range buildings {
		path := gp.PathResolver.Resolve("building", b)
		gp.BuildingTextures[b] = rl.LoadTexture(path)
	}
}

func (gp *GamePage) loadEnvironmentObjects() {
	type AssetSpec struct {
		SubFolder   string
		FileName    string
		CustomScale float32
	}

	specs := []AssetSpec{
		{"building", "Tent.png", 0.6},
		{"building", "Red-Banner.png", 0.5},
		{"building", "Well.png", 0.6},
		{"building", "Wooden-Bridge-Horizontal.png", 0.8},
		{"building", "Campfire.png", 0.5},
		{"forest", "Bushes-Large.png", 0.3},
		{"forest", "Tree-Large.png", 0.4},
	}

	gp.SceneryObjects = make([]EnvObjectInfo, len(specs))
	for i, spec := range specs {
		fullPath := gp.PathResolver.Resolve(spec.SubFolder, spec.FileName)
		gp.SceneryObjects[i] = EnvObjectInfo{
			Path:     fullPath,
			Category: CategoryScenery,
			Scale:    spec.CustomScale,
			Texture:  rl.LoadTexture(fullPath),
		}
	}
}

func (gp *GamePage) Unload() {
	if gp.IsUnloaded {
		return
	}
	for _, tex := range gp.PathTextures {
		if tex.ID > 0 {
			rl.UnloadTexture(tex)
		}
	}
	for _, tex := range gp.BuildingTextures {
		if tex.ID > 0 {
			rl.UnloadTexture(tex)
		}
	}
	if gp.EmptyTexture.ID > 0 {
		rl.UnloadTexture(gp.EmptyTexture)
	}
	for _, obj := range gp.SceneryObjects {
		if obj.Texture.ID > 0 {
			rl.UnloadTexture(obj.Texture)
		}
	}
	if gp.MainChar != nil {
		gp.MainChar.UnloadAnimations()
	}
	for _, e := range gp.Enemies {
		e.UnloadAnimations()
	}
	for _, n := range gp.NPCs {
		n.UnloadAnimations()
	}

	gp.BasePage.Unload()
}

func (gp *GamePage) EventListener() {
	if rl.IsKeyPressed(rl.KeyEscape) {
		gp.NextState = enums.MainMenu
	}
}

func updateMonsterAI(m *component.Monster, dt float32, target rl.Vector2, world *WorldMap) {
	if m.IsDead {
		return
	}
	dist := rl.Vector2Distance(m.Position, target)

	if dist < TileSize*5 && dist > TileSize*0.5 {
		m.SetAction(component.ActionWalking)
		dir := rl.Vector2Subtract(target, m.Position)
		length := rl.Vector2Length(dir)
		if length > 0 {
			dir.X /= length
			dir.Y /= length
		}

		if math.Abs(float64(dir.X)) > math.Abs(float64(dir.Y)) {
			if dir.X > 0 {
				m.Direction = component.DirRight
			} else {
				m.Direction = component.DirLeft
			}
		} else {
			if dir.Y > 0 {
				m.Direction = component.DirFront
			} else {
				m.Direction = component.DirBack
			}
		}

		nextX := m.Position.X + dir.X*m.Speed*0.4*dt
		nextY := m.Position.Y + dir.Y*m.Speed*0.4*dt
		if isWalkableWorld(nextX, m.Position.Y, world) {
			m.Position.X = nextX
		}
		if isWalkableWorld(m.Position.X, nextY, world) {
			m.Position.Y = nextY
		}
	} else {
		m.SetAction(component.ActionIdle)
	}
	m.AdvanceFrame(dt, true)
}

func (gp *GamePage) Update(dt float32) {
	if gp.MainChar.IsDead {
		gp.MainChar.AdvanceFrame(dt, false)
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
		if isWalkableWorld(nextX-radius, gp.MainChar.Position.Y-radius, gp.World) &&
			isWalkableWorld(nextX+radius, gp.MainChar.Position.Y+radius, gp.World) {
			gp.MainChar.Position.X = nextX
		}
		if isWalkableWorld(gp.MainChar.Position.X-radius, nextY-radius, gp.World) &&
			isWalkableWorld(gp.MainChar.Position.X+radius, nextY+radius, gp.World) {
			gp.MainChar.Position.Y = nextY
		}
	} else {
		gp.MainChar.SetAction(component.ActionIdle)
	}

	for _, e := range gp.Enemies {
		updateMonsterAI(e, dt, gp.MainChar.Position, gp.World)
	}
	for _, n := range gp.NPCs {
		n.AdvanceFrame(dt, true)
	}

	gp.Camera.Target = gp.MainChar.Position

	charTileX := int(gp.MainChar.Position.X / TileSize)
	charTileY := int(gp.MainChar.Position.Y / TileSize)
	for _, room := range gp.World.Rooms {
		if abs(charTileX-room.TileX) <= 3 && abs(charTileY-room.TileY) <= 3 {
			gp.ActiveRoom = room
			break
		}
	}

	gp.MainChar.AdvanceFrame(dt, true)
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func isWalkableWorld(x, y float32, w *WorldMap) bool {
	gridX := int(x / TileSize)
	gridY := int(y / TileSize)
	if gridX < 0 || gridX >= w.Width || gridY < 0 || gridY >= w.Height {
		return false
	}
	tile := w.Grid[gridY][gridX]
	return tile == TilePath || tile == TileRoom
}

func (gp *GamePage) Render() {
	gp.EventListener()
	dt := rl.GetFrameTime()
	gp.Update(dt)

	rl.ClearBackground(rl.NewColor(15, 20, 15, 255))
	rl.BeginMode2D(gp.Camera)

	for y := 0; y < gp.World.Height; y++ {
		for x := 0; x < gp.World.Width; x++ {
			pos := rl.NewVector2(float32(x)*TileSize, float32(y)*TileSize)
			destRect := rl.NewRectangle(pos.X, pos.Y, TileSize, TileSize)
			tile := gp.World.Grid[y][x]

			switch tile {
			case TilePath:
				mask := getPathBitmaskWorld(gp.World, x, y)
				tex, exists := gp.PathTextures[mask]
				if exists && tex.ID > 0 {
					srcRect := rl.NewRectangle(0, 0, float32(tex.Width), float32(tex.Height))
					rl.DrawTexturePro(tex, srcRect, destRect, rl.Vector2Zero(), 0, rl.White)
				} else {
					rl.DrawRectangleRec(destRect, rl.NewColor(160, 120, 80, 255))
				}

			case TileRoom:
				if gp.EmptyTexture.ID > 0 {
					srcEmpty := rl.NewRectangle(0, 0, float32(gp.EmptyTexture.Width), float32(gp.EmptyTexture.Height))
					rl.DrawTexturePro(gp.EmptyTexture, srcEmpty, destRect, rl.Vector2Zero(), 0, rl.White)
				} else {
					rl.DrawRectangleRec(destRect, rl.NewColor(200, 170, 120, 255))
				}

			case TileWall:
				if gp.EmptyTexture.ID > 0 {
					srcEmpty := rl.NewRectangle(0, 0, float32(gp.EmptyTexture.Width), float32(gp.EmptyTexture.Height))
					rl.DrawTexturePro(gp.EmptyTexture, srcEmpty, destRect, rl.Vector2Zero(), 0, rl.White)
				} else {
					rl.DrawRectangleRec(destRect, rl.NewColor(34, 139, 34, 255))
				}

				idx := gp.World.EnvObjectIdx[y][x]
				if idx >= 0 && idx < len(gp.SceneryObjects) {
					obj := gp.SceneryObjects[idx]
					if obj.Texture.ID > 0 {
						srcRect := rl.NewRectangle(0, 0, float32(obj.Texture.Width), float32(obj.Texture.Height))
						scale := (TileSize / float32(obj.Texture.Width)) * obj.Scale * 3.0
						drawW := float32(obj.Texture.Width) * scale
						drawH := float32(obj.Texture.Height) * scale
						drawRect := rl.NewRectangle(pos.X+TileSize/2-drawW/2, pos.Y+TileSize/2-drawH/2, drawW, drawH)
						rl.DrawTexturePro(obj.Texture, srcRect, drawRect, rl.Vector2Zero(), 0, rl.White)
					}
				}
			}
		}
	}

	for _, room := range gp.World.Rooms {
		pos := rl.NewVector2(float32(room.TileX)*TileSize, float32(room.TileY)*TileSize)
		bldSize := TileSize * 4.0

		var tex rl.Texture2D
		nameLower := strings.ToLower(room.Name)
		if strings.Contains(nameLower, "shop") {
			tex = gp.BuildingTextures["shop.png"]
		} else if strings.Contains(nameLower, "tavern") {
			tex = gp.BuildingTextures["Tavern.png"]
		} else if strings.Contains(nameLower, "boss") {
			tex = gp.BuildingTextures["Castle-Round.png"]
		} else if strings.Contains(nameLower, "cave") {
			tex = gp.BuildingTextures["Tent.png"]
		} else {
			tex = gp.BuildingTextures["House.png"]
		}

		if tex.ID > 0 {
			scale := bldSize / float32(tex.Width)
			destRect := rl.NewRectangle(pos.X-bldSize/2, pos.Y-bldSize/2, float32(tex.Width)*scale, float32(tex.Height)*scale)
			rl.DrawTexturePro(tex, rl.NewRectangle(0, 0, float32(tex.Width), float32(tex.Height)), destRect, rl.Vector2Zero(), 0, rl.White)
		} else {
			rl.DrawRectangle(int32(pos.X-bldSize/2), int32(pos.Y-bldSize/2), int32(bldSize), int32(bldSize), rl.DarkGray)
		}
		rl.DrawText(room.Name, int32(pos.X-50), int32(pos.Y-bldSize/2-30), 20, rl.RayWhite)
	}

	// Render imported components native methods
	for _, n := range gp.NPCs {
		n.Draw()
	}
	for _, e := range gp.Enemies {
		e.Draw()
	}
	if gp.MainChar != nil {
		gp.MainChar.Draw()
	}

	rl.EndMode2D()

	// Screen Space HUD
	rl.DrawRectangle(10, 10, 480, 70, rl.NewColor(0, 0, 0, 180))
	rl.DrawRectangleLines(10, 10, 480, 70, rl.RayWhite)
	rl.DrawText("Use WASD/Arrow Keys to Move | ESC to Menu", 20, 18, 14, rl.RayWhite)

	if gp.ActiveRoom != nil {
		infoText := fmt.Sprintf("Current Room: %s", gp.ActiveRoom.Name)
		rl.DrawText(infoText, 20, 38, 16, rl.Gold)
		rl.DrawText(gp.ActiveRoom.Description, 20, 56, 12, rl.LightGray)
	}
}

func getPathBitmaskWorld(w *WorldMap, x, y int) string {
	north, east, south, west := 0, 0, 0, 0
	if y > 0 && (w.Grid[y-1][x] == TilePath || w.Grid[y-1][x] == TileRoom) {
		north = 1
	}
	if x < w.Width-1 && (w.Grid[y][x+1] == TilePath || w.Grid[y][x+1] == TileRoom) {
		east = 1
	}
	if y < w.Height-1 && (w.Grid[y+1][x] == TilePath || w.Grid[y+1][x] == TileRoom) {
		south = 1
	}
	if x > 0 && (w.Grid[y][x-1] == TilePath || w.Grid[y][x-1] == TileRoom) {
		west = 1
	}
	return fmt.Sprintf("%d%d%d%d", north, east, south, west)
}