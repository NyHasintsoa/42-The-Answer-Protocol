package service

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"tap-gui/src/utils"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	BaseTileSize = 48
	GroundScale  = 1.0
	TileSize     = float32(BaseTileSize) * GroundScale
	CharScale    = 0.2
	CellSpacing  = 12
)

type TileType int

const (
	TileWall TileType = iota
	TilePath
	TileRoom
)

type EnvCategory int

const (
	CategoryScenery EnvCategory = iota
)

type EnvObjectInfo struct {
	Path     string
	Category EnvCategory
	Scale    float32
	Texture  rl.Texture2D
}

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

type Monster interface {
	UpdateAI(dt float32, target rl.Vector2, isWalkable func(float32, float32) bool)
	Render()
	UnloadAnimations()
}

type INpc interface {
	Update(dt float32)
	Render()
	UnloadAnimations()
}

type EntityFactory struct {
	NewMonster func(x, y, scale float32, resolver *utils.PathResolver, assetPath, name string) Monster
	NewNPC     func(kind, relativeDir string, x, y, scale float32) (INpc, error)
}

type MapManager struct {
	World          *WorldMap
	Enemies        []Monster
	NPCs           []INpc
	PathTextures   map[string]rl.Texture2D
	RoomTextures   map[string]rl.Texture2D
	EmptyTexture   rl.Texture2D
	SceneryObjects []EnvObjectInfo
	PathResolver   *utils.PathResolver
	ActiveRoom     *RoomNode
	entityFactory  EntityFactory
}

func NewMapManager(pathResolver *utils.PathResolver, configJSON string, factory EntityFactory) *MapManager {
	mm := &MapManager{
		PathResolver:  pathResolver,
		PathTextures:  make(map[string]rl.Texture2D),
		RoomTextures:  make(map[string]rl.Texture2D),
		Enemies:       make([]Monster, 0),
		NPCs:          make([]INpc, 0),
		entityFactory: factory,
	}

	mm.loadGroundTextures()
	mm.loadEnvironmentObjects()

	mm.World = mm.generateWorldMap(configJSON)
	mm.spawnEntities()

	return mm
}

func (mm *MapManager) GetStartPosition() (float32, float32) {
	if startRoom, exists := mm.World.Rooms["start"]; exists {
		mm.ActiveRoom = startRoom
		return float32(startRoom.TileX)*TileSize + TileSize*2, float32(startRoom.TileY)*TileSize + TileSize*2
	}
	return TileSize, TileSize
}

func (mm *MapManager) IsWalkable(x, y float32) bool {
	gridX := int(x / TileSize)
	gridY := int(y / TileSize)
	if gridX < 0 || gridX >= mm.World.Width || gridY < 0 || gridY >= mm.World.Height {
		return false
	}
	tile := mm.World.Grid[gridY][gridX]
	return tile == TilePath || tile == TileRoom
}

func (mm *MapManager) Update(dt float32, playerPos rl.Vector2) {
	for _, n := range mm.NPCs {
		n.Update(dt)
	}

	charTileX := int(playerPos.X / TileSize)
	charTileY := int(playerPos.Y / TileSize)
	for _, room := range mm.World.Rooms {
		if charTileX >= room.TileX && charTileX < room.TileX+4 &&
			charTileY >= room.TileY && charTileY < room.TileY+4 {
			mm.ActiveRoom = room
			break
		}
	}
}

func (mm *MapManager) GetPathTileTexture(x, y int) (rl.Texture2D, bool) {
	mask := 0
	if y > 0 && (mm.World.Grid[y-1][x] == TilePath || mm.World.Grid[y-1][x] == TileRoom) {
		mask |= 8
	}
	if x < mm.World.Width-1 && (mm.World.Grid[y][x+1] == TilePath || mm.World.Grid[y][x+1] == TileRoom) {
		mask |= 4
	}
	if y < mm.World.Height-1 && (mm.World.Grid[y+1][x] == TilePath || mm.World.Grid[y+1][x] == TileRoom) {
		mask |= 2
	}
	if x > 0 && (mm.World.Grid[y][x-1] == TilePath || mm.World.Grid[y][x-1] == TileRoom) {
		mask |= 1
	}

	key := fmt.Sprintf("%04b", mask)
	if tex, ok := mm.PathTextures[key]; ok && tex.ID > 0 {
		return tex, true
	}
	if tex, ok := mm.PathTextures["1111"]; ok && tex.ID > 0 {
		return tex, true
	}
	return rl.Texture2D{}, false
}

func (mm *MapManager) GetRoomTileTexture(x, y int) (rl.Texture2D, bool) {
	for _, room := range mm.World.Rooms {
		dx := x - room.TileX
		dy := y - room.TileY
		if dx >= 0 && dx < 4 && dy >= 0 && dy < 4 {
			if tex, ok := mm.RoomTextures["full"]; ok && tex.ID > 0 {
				return tex, true
			}
		}
	}
	return rl.Texture2D{}, false
}

func (mm *MapManager) Unload() {
	for _, e := range mm.Enemies {
		e.UnloadAnimations()
	}
	for _, n := range mm.NPCs {
		n.UnloadAnimations()
	}
}

func (mm *MapManager) spawnEntities() {
	for _, room := range mm.World.Rooms {
		roomX := float32(room.TileX)*TileSize + TileSize/2
		roomY := float32(room.TileY)*TileSize + TileSize/2

		for _, npcName := range room.NPCs {
			nameLower := strings.ToLower(npcName)
			if strings.Contains(nameLower, "boss") {
				if mm.entityFactory.NewMonster != nil {
					e := mm.entityFactory.NewMonster(roomX+100, roomY+100, CharScale, mm.PathResolver, mm.PathResolver.Resolve("characters", "boss"), "Boss")
					mm.Enemies = append(mm.Enemies, e)
				}
			} else if strings.Contains(nameLower, "goblin") {
				if mm.entityFactory.NewMonster != nil {
					e := mm.entityFactory.NewMonster(roomX+100, roomY-100, CharScale, mm.PathResolver, mm.PathResolver.Resolve("characters", "gobelin"), "Goblin")
					mm.Enemies = append(mm.Enemies, e)
				}
			} else if strings.Contains(nameLower, "guard") {
				relDir := filepath.Join("npc", "guard")
				if mm.entityFactory.NewNPC != nil {
					if n, err := mm.entityFactory.NewNPC("guard", relDir, roomX-120, roomY+80, CharScale); err == nil {
						mm.NPCs = append(mm.NPCs, n)
					}
				}
			} else if strings.Contains(nameLower, "seller") || strings.Contains(nameLower, "shop") {
				relDir := filepath.Join("npc", "seller")
				if mm.entityFactory.NewNPC != nil {
					if n, err := mm.entityFactory.NewNPC("seller", relDir, roomX-120, roomY+80, CharScale); err == nil {
						mm.NPCs = append(mm.NPCs, n)
					}
				}
			} else {
				relDir := filepath.Join("npc", nameLower)
				if mm.entityFactory.NewNPC != nil {
					if n, err := mm.entityFactory.NewNPC("person", relDir, roomX-120, roomY+80, CharScale); err != nil {
						relDir = filepath.Join("npc", "person")
						if n, err = mm.entityFactory.NewNPC("person", relDir, roomX-120, roomY+80, CharScale); err == nil {
							mm.NPCs = append(mm.NPCs, n)
						}
					} else {
						mm.NPCs = append(mm.NPCs, n)
					}
				}
			}
		}
	}
}

func (mm *MapManager) generateWorldMap(configJSON string) *WorldMap {
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

	padding := 10
	mapWidth := (maxX-minX+1)*CellSpacing + padding*2
	mapHeight := (maxY-minY+1)*CellSpacing + padding*2

	grid := make([][]TileType, mapHeight)
	envIdx := make([][]int, mapHeight)

	for y := 0; y < mapHeight; y++ {
		grid[y] = make([]TileType, mapWidth)
		envIdx[y] = make([]int, mapWidth)
		for x := 0; x < mapWidth; x++ {
			envIdx[y][x] = -1
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
		tileX := (coord[0]-minX)*CellSpacing + padding
		tileY := (coord[1]-minY)*CellSpacing + padding

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

		// 4x4 room sizing
		for ry := 0; ry < 4; ry++ {
			for rx := 0; rx < 4; rx++ {
				if tileY+ry >= 0 && tileY+ry < mapHeight && tileX+rx >= 0 && tileX+rx < mapWidth {
					grid[tileY+ry][tileX+rx] = TileRoom
				}
			}
		}
	}

	// 2x2 wide paths centered on the 4x4 room edges
	for roomID, roomNode := range worldMap.Rooms {
		roomCfg := config.World.Rooms[roomID]
		for dir, targetID := range roomCfg.Exits {
			if targetNode, exists := worldMap.Rooms[targetID]; exists {
				switch dir {
				case "north":
					mm.carvePath(grid, roomNode.TileX+1, roomNode.TileY, targetNode.TileX+1, targetNode.TileY+3)
					mm.carvePath(grid, roomNode.TileX+2, roomNode.TileY, targetNode.TileX+2, targetNode.TileY+3)
				case "south":
					mm.carvePath(grid, roomNode.TileX+1, roomNode.TileY+3, targetNode.TileX+1, targetNode.TileY)
					mm.carvePath(grid, roomNode.TileX+2, roomNode.TileY+3, targetNode.TileX+2, targetNode.TileY)
				case "east":
					mm.carvePath(grid, roomNode.TileX+3, roomNode.TileY+1, targetNode.TileX, targetNode.TileY+1)
					mm.carvePath(grid, roomNode.TileX+3, roomNode.TileY+2, targetNode.TileX, targetNode.TileY+2)
				case "west":
					mm.carvePath(grid, roomNode.TileX, roomNode.TileY+1, targetNode.TileX+3, targetNode.TileY+1)
					mm.carvePath(grid, roomNode.TileX, roomNode.TileY+2, targetNode.TileX+3, targetNode.TileY+2)
				}
			}
		}
	}

	for y := 0; y < mapHeight; y++ {
		for x := 0; x < mapWidth; x++ {
			if grid[y][x] == TileWall {
				val := (x*31 + y*17 + (x*y)*7) % 100
				if val < 45 && len(mm.SceneryObjects) > 0 {
					envIdx[y][x] = (x*13 + y*7) % len(mm.SceneryObjects)
				}
			}
		}
	}

	return worldMap
}

func (mm *MapManager) carvePath(grid [][]TileType, x1, y1, x2, y2 int) {
	currX, currY := x1, y1
	for {
		if currY >= 0 && currY < len(grid) && currX >= 0 && currX < len(grid[0]) {
			if grid[currY][currX] == TileWall {
				grid[currY][currX] = TilePath
			}
		}
		if currX == x2 && currY == y2 {
			break
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

func (mm *MapManager) loadGroundTextures() {
	for i := range 16 {
		key := fmt.Sprintf("%04b", i)
		path := mm.PathResolver.Resolve("ground", key+".png")
		mm.PathTextures[key] = mm.PathResolver.ImageManager.Load(path)
	}
	mm.EmptyTexture = mm.PathResolver.ImageManager.Load(mm.PathResolver.Resolve("ground", "lawn.png"))

	fullPath := mm.PathResolver.Resolve("ground", "full.png")
	mm.RoomTextures["full"] = mm.PathResolver.ImageManager.Load(fullPath)
}

func (mm *MapManager) loadEnvironmentObjects() {
	forestFiles := []struct {
		FileName string
		Scale    float32
	}{
		{"Tree-Large.png", 0.45},
		{"Tree-Medium.png", 0.4},
		{"Tree-Small.png", 0.35},
		{"Bushes-Large.png", 0.3},
		{"Bushes-Medium.png", 0.25},
		{"Bushes-Small.png", 0.2},
		{"Rock-01.png", 0.3},
		{"Rock-02.png", 0.3},
		{"Rock-03.png", 0.3},
		{"Rock-04.png", 0.3},
		{"Rock-05.png", 0.3},
		{"Tree-Stump-Short.png", 0.3},
		{"Tree-Stump-Tall.png", 0.35},
	}

	mm.SceneryObjects = make([]EnvObjectInfo, len(forestFiles))
	for i, file := range forestFiles {
		fullPath := mm.PathResolver.Resolve("forest", file.FileName)
		mm.SceneryObjects[i] = EnvObjectInfo{
			Path:     fullPath,
			Category: CategoryScenery,
			Scale:    file.Scale,
			Texture:  mm.PathResolver.ImageManager.Load(fullPath),
		}
	}
}

func (mm *MapManager) abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
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