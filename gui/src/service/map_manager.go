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
	BaseTileSize = 64
	GroundScale  = 1.0
	TileSize     = float32(BaseTileSize) * GroundScale
	CharScale    = 0.25
	CellSpacing  = 15
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
	CategoryBuilding
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
	World            *WorldMap
	Enemies          []Monster
	NPCs             []INpc
	PathTextures     map[string]rl.Texture2D
	BuildingTextures map[string]rl.Texture2D
	EmptyTexture     rl.Texture2D
	SceneryObjects   []EnvObjectInfo
	PathResolver     *utils.PathResolver
	ActiveRoom       *RoomNode
	entityFactory    EntityFactory
}

func NewMapManager(pathResolver *utils.PathResolver, configJSON string, factory EntityFactory) *MapManager {
	mm := &MapManager{
		PathResolver:  pathResolver,
		PathTextures:  make(map[string]rl.Texture2D),
		Enemies:       make([]Monster, 0),
		NPCs:          make([]INpc, 0),
		entityFactory: factory,
	}

	mm.loadGroundTextures()
	mm.loadBuildingTextures()
	mm.loadEnvironmentObjects()

	mm.World = mm.generateWorldMap(configJSON)
	mm.spawnEntities()

	return mm
}

func (mm *MapManager) GetStartPosition() (float32, float32) {
	if startRoom, exists := mm.World.Rooms["start"]; exists {
		mm.ActiveRoom = startRoom
		return float32(startRoom.TileX)*TileSize + TileSize/2, float32(startRoom.TileY)*TileSize + TileSize/2
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
	for _, e := range mm.Enemies {
		e.UpdateAI(dt, playerPos, mm.IsWalkable)
	}
	for _, n := range mm.NPCs {
		n.Update(dt)
	}

	charTileX := int(playerPos.X / TileSize)
	charTileY := int(playerPos.Y / TileSize)
	for _, room := range mm.World.Rooms {
		if mm.abs(charTileX-room.TileX) <= 3 && mm.abs(charTileY-room.TileY) <= 3 {
			mm.ActiveRoom = room
			break
		}
	}
}

func (mm *MapManager) GetPathBitmaskWorld(x, y int) string {
	north, east, south, west := 0, 0, 0, 0
	if y > 0 && (mm.World.Grid[y-1][x] == TilePath || mm.World.Grid[y-1][x] == TileRoom) {
		north = 1
	}
	if x < mm.World.Width-1 && (mm.World.Grid[y][x+1] == TilePath || mm.World.Grid[y][x+1] == TileRoom) {
		east = 1
	}
	if y < mm.World.Height-1 && (mm.World.Grid[y+1][x] == TilePath || mm.World.Grid[y+1][x] == TileRoom) {
		south = 1
	}
	if x > 0 && (mm.World.Grid[y][x-1] == TilePath || mm.World.Grid[y][x-1] == TileRoom) {
		west = 1
	}
	return fmt.Sprintf("%d%d%d%d", north, east, south, west)
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

	padding := 20
	mapWidth := (maxX-minX+1)*CellSpacing + padding*2
	mapHeight := (maxY-minY+1)*CellSpacing + padding*2

	grid := make([][]TileType, mapHeight)
	envIdx := make([][]int, mapHeight)

	for y := 0; y < mapHeight; y++ {
		grid[y] = make([]TileType, mapWidth)
		envIdx[y] = make([]int, mapWidth)
		for x := 0; x < mapWidth; x++ {
			if (x*17+y*11)%10 >= 4 {
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
				mm.carvePath(grid, roomNode.TileX, roomNode.TileY, targetNode.TileX, targetNode.TileY)
			}
		}
	}

	return worldMap
}

func (mm *MapManager) carvePath(grid [][]TileType, x1, y1, x2, y2 int) {
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

func (mm *MapManager) loadGroundTextures() {
	bitmasks := []string{
		"0000", "0001", "0010", "0011",
		"0100", "0101", "0110", "0111",
		"1000", "1001", "1010", "1011",
		"1100", "1101", "1110", "1111",
	}
	for _, mask := range bitmasks {
		path := mm.PathResolver.Resolve("ground", mask+".png")
		mm.PathTextures[mask] = mm.PathResolver.ImageManager.Load(path)
	}
	mm.EmptyTexture = mm.PathResolver.ImageManager.Load(mm.PathResolver.Resolve("ground", "lawn.png"))
}

func (mm *MapManager) loadBuildingTextures() {
	mm.BuildingTextures = make(map[string]rl.Texture2D)
	buildings := []string{"House.png", "shop.png", "Tavern.png", "Castle-Round.png", "Tent.png"}
	for _, b := range buildings {
		path := mm.PathResolver.Resolve("building", b)
		mm.BuildingTextures[b] = mm.PathResolver.ImageManager.Load(path)
	}
}

func (mm *MapManager) loadEnvironmentObjects() {
	specs := []struct {
		SubFolder   string
		FileName    string
		CustomScale float32
	}{
		{"building", "Tent.png", 0.6},
		{"building", "Red-Banner.png", 0.5},
		{"building", "Well.png", 0.6},
		{"building", "Wooden-Bridge-Horizontal.png", 0.8},
		{"building", "Campfire.png", 0.5},
		{"forest", "Bushes-Large.png", 0.3},
		{"forest", "Tree-Large.png", 0.4},
	}

	mm.SceneryObjects = make([]EnvObjectInfo, len(specs))
	for i, spec := range specs {
		fullPath := mm.PathResolver.Resolve(spec.SubFolder, spec.FileName)
		mm.SceneryObjects[i] = EnvObjectInfo{
			Path:     fullPath,
			Category: CategoryScenery,
			Scale:    spec.CustomScale,
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
