package service

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"tap-gui/src/model/npc"
	"tap-gui/src/utils"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	// BaseTileSize is the pixel size of a ground tile before scaling.
	BaseTileSize = 48
	// GroundScale is the scale applied to ground tiles.
	GroundScale = 1.0
	// TileSize is the rendered size of one ground tile in world units.
	TileSize = float32(BaseTileSize) * GroundScale
	// CharScale is the rendering scale applied to map characters.
	CharScale = 0.2
	// CellSpacing is the number of grid cells between room origins.
	CellSpacing = 12
)

// TileType identifies the kind of terrain occupying a map cell.
type TileType int

const (
	// TileWall marks an impassable map cell.
	TileWall TileType = iota
	// TilePath marks a traversable path cell.
	TilePath
	// TileRoom marks a cell inside a room.
	TileRoom
)

// EnvCategory identifies the type of environmental object drawn on the map.
type EnvCategory int

const (
	// CategoryScenery identifies decorative environmental objects.
	CategoryScenery EnvCategory = iota
)

// EnvObjectInfo describes a scenery asset and its rendering properties.
type EnvObjectInfo struct {
	Path     string
	Category EnvCategory
	Scale    float32
	Texture  rl.Texture2D
}

// RoomConfig contains server-provided data for one world room.
type RoomConfig struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Exits       map[string]string `json:"exits"`
	Items       []string          `json:"items"`
	NPCs        []string          `json:"npcs"`
}

// MapConfig contains the world and entity metadata received from the server.
type MapConfig struct {
	World struct {
		Rooms map[string]RoomConfig `json:"rooms"`
	} `json:"world"`
	NPCs   map[string]npc.NPCConfig `json:"npcs"`
	Items  map[string]ItemConfig    `json:"items"`
	Quests map[string]QuestConfig   `json:"quests"`
}

// ItemConfig contains display metadata for a server-defined item.
type ItemConfig struct {
	Name        string `json:"name"`
	Cost        int    `json:"cost"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

// QuestConfig contains display metadata for a server-defined quest.
type QuestConfig struct {
	Title       string `json:"title"`
	Giver       string `json:"giver"`
	Description string `json:"description"`
	Reward      string `json:"reward"`
}

// ValidateMapConfigJSON decodes configJSON and checks its required room data.
func ValidateMapConfigJSON(configJSON string) error {
	var config MapConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return fmt.Errorf("decode map config: %w", err)
	}
	if len(config.World.Rooms) == 0 {
		return fmt.Errorf("map config has no world rooms")
	}
	if _, exists := config.World.Rooms["start"]; !exists {
		return fmt.Errorf("map config has no start room")
	}
	for roomID, room := range config.World.Rooms {
		if room.Name == "" {
			return fmt.Errorf("map config room %q has no name", roomID)
		}
		for direction, destination := range room.Exits {
			if _, exists := config.World.Rooms[destination]; !exists {
				return fmt.Errorf("map config room %q exits %q to unknown room %q", roomID, direction, destination)
			}
		}
	}
	return nil
}

// RoomNode stores a room's world position and server-provided contents.
type RoomNode struct {
	ID          string
	Name        string
	Description string
	Exits       map[string]string
	GridX       int
	GridY       int
	TileX       int
	TileY       int
	Color       rl.Color
	NPCs        []string
	Items       []string
}

// WorldMap stores the generated tile grid and its room graph.
type WorldMap struct {
	Width        int
	Height       int
	Grid         [][]TileType
	Rooms        map[string]*RoomNode
	EnvObjectIdx [][]int
}

// MapManager builds and updates the graphical world from server map data.
type MapManager struct {
	World          *WorldMap
	NPCConfigs     map[string]npc.NPCConfig
	ItemConfigs    map[string]ItemConfig
	QuestConfigs   map[string]QuestConfig
	Enemies        []npc.Monster
	NPCs           []npc.INpc
	PathTextures   map[string]rl.Texture2D
	RoomTextures   map[string]rl.Texture2D
	EmptyTexture   rl.Texture2D
	SceneryObjects []EnvObjectInfo
	PathResolver   *utils.PathResolver
	ActiveRoom     *RoomNode
	entityFactory  npc.EntityFactory
}

// NewMapManager loads map assets, creates the world described by configJSON, and spawns its entities.
// pathResolver locates assets; configJSON contains world and NPC configuration; factory creates
// the runtime NPC and monster entities.
func NewMapManager(pathResolver *utils.PathResolver, configJSON string, factory npc.EntityFactory) *MapManager {
	mm := &MapManager{
		PathResolver:  pathResolver,
		PathTextures:  make(map[string]rl.Texture2D),
		RoomTextures:  make(map[string]rl.Texture2D),
		Enemies:       make([]npc.Monster, 0),
		NPCs:          make([]npc.INpc, 0),
		entityFactory: factory,
	}

	mm.loadGroundTextures()
	mm.loadEnvironmentObjects()

	mm.World = mm.generateWorldMap(configJSON)
	mm.spawnEntities()

	return mm
}

// GetStartPosition returns the center of the start room, or a default map position if it is absent.
func (mm *MapManager) GetStartPosition() (float32, float32) {
	if startRoom, exists := mm.World.Rooms["start"]; exists {
		mm.ActiveRoom = startRoom
		return float32(startRoom.TileX)*TileSize + TileSize*2, float32(startRoom.TileY)*TileSize + TileSize*2
	}
	return TileSize, TileSize
}

// IsWalkable reports whether the world-space position x, y is on a path or room tile.
// x and y are the horizontal and vertical world coordinates.
func (mm *MapManager) IsWalkable(x, y float32) bool {
	gridX := int(x / TileSize)
	gridY := int(y / TileSize)
	if gridX < 0 || gridX >= mm.World.Width || gridY < 0 || gridY >= mm.World.Height {
		return false
	}
	tile := mm.World.Grid[gridY][gridX]
	return tile == TilePath || tile == TileRoom
}

// GetRoomInDirection returns the adjacent room connected through direction.
// direction is an exit name such as "north"; the boolean is false if there is no matching room.
func (mm *MapManager) GetRoomInDirection(direction string) (*RoomNode, bool) {
	if mm.ActiveRoom == nil {
		return nil, false
	}

	targetID, exists := mm.ActiveRoom.Exits[strings.ToLower(direction)]
	if !exists {
		return nil, false
	}

	room, exists := mm.World.Rooms[targetID]
	if !exists {
		return nil, false
	}

	return room, true
}

// SetActiveRoom selects the server-reported room and reports whether it exists in the configured world.
func (mm *MapManager) SetActiveRoom(roomID string) bool {
	room, exists := mm.World.Rooms[roomID]
	if !exists {
		return false
	}
	mm.ActiveRoom = room
	return true
}

// GetNPCConfig returns the NPC configuration associated with key.
// key identifies an NPC; the boolean reports whether a configuration was found.
func (mm *MapManager) GetNPCConfig(key string) (npc.NPCConfig, bool) {
	if cfg, exists := mm.NPCConfigs[key]; exists {
		if cfg.Key == "" {
			cfg.Key = key
		}
		return cfg, true
	}
	return npc.NPCConfig{Key: key, Name: key}, false
}

// GetItemConfig returns the server-defined metadata for itemID.
func (mm *MapManager) GetItemConfig(itemID string) (ItemConfig, bool) {
	item, exists := mm.ItemConfigs[itemID]
	return item, exists
}

// GetQuestConfig returns the server-defined metadata for questID.
func (mm *MapManager) GetQuestConfig(questID string) (QuestConfig, bool) {
	quest, exists := mm.QuestConfigs[questID]
	return quest, exists
}

// Update advances NPC animations and sets the active room containing playerPos.
// dt is elapsed time in seconds; playerPos is the player's world-space position.
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

// GetPathTileTexture returns the path texture matching the neighboring tiles at grid position x, y.
// x and y are tile coordinates; the boolean reports whether a usable texture was found.
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

// GetRoomTileTexture returns the room texture at grid position x, y when one is available.
// x and y are tile coordinates; the boolean reports whether a usable texture was found.
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

// Unload clears map entities, textures, and world state owned by the manager.
func (mm *MapManager) Unload() {
	if mm == nil {
		return
	}
	for _, e := range mm.Enemies {
		e.UnloadAnimations()
	}
	for _, n := range mm.NPCs {
		n.UnloadAnimations()
	}
	mm.Enemies = nil
	mm.NPCs = nil
	mm.PathTextures = nil
	mm.RoomTextures = nil
	mm.EmptyTexture = rl.Texture2D{}
	mm.SceneryObjects = nil
	mm.NPCConfigs = nil
	mm.ActiveRoom = nil
	mm.World = nil
}

// spawnEntities creates and positions the NPCs and monsters defined in each room.
func (mm *MapManager) spawnEntities() {
	mm.Enemies = make([]npc.Monster, 0)
	mm.NPCs = make([]npc.INpc, 0)

	for _, room := range mm.World.Rooms {
		roomX := float32(room.TileX)*TileSize + TileSize*2
		roomY := float32(room.TileY)*TileSize + TileSize*2

		totalEntities := len(room.NPCs)
		for i, npcKey := range room.NPCs {
			offsetX := float32(i-totalEntities/2) * 32.0
			spawnX := roomX + offsetX
			spawnY := roomY

			npcDef, exists := mm.GetNPCConfig(npcKey)

			var imageType string
			var npcType string
			npcName := npcKey

			if exists {
				imageType = strings.ToLower(npcDef.ImageType)
				npcType = strings.ToLower(npcDef.Type)
				if npcDef.Name != "" {
					npcName = npcDef.Name
				}
			} else {
				imageType = "person"
			}

			if imageType == "boss" || imageType == "gobelin" || npcType == "boss" || npcType == "gobelin" {
				charFolder := imageType
				if mm.entityFactory.NewMonster != nil {
					assetPath := mm.PathResolver.Resolve("characters", charFolder)
					e := mm.entityFactory.NewMonster(CharScale, mm.PathResolver, assetPath, npcName)
					if e != nil {
						e.SetPosition(spawnX, spawnY)
						mm.Enemies = append(mm.Enemies, e)
					}
				}
			} else {
				kind := "person"
				switch imageType {
				case "guard":
					kind = "guard"
				case "merchant":
					kind = "merchant"
				}

				relDir := filepath.Join("npc", imageType)
				if mm.entityFactory.NewNPC != nil {
					n, err := mm.entityFactory.NewNPC(kind, relDir, CharScale)
					if err != nil {
						relDir = filepath.Join("npc", "person")
						n, err = mm.entityFactory.NewNPC("person", relDir, CharScale)
					}
					if err == nil && n != nil {
						n.SetPosition(spawnX, spawnY)
						mm.NPCs = append(mm.NPCs, n)
					}
				}
			}
		}
	}
}

// generateWorldMap decodes the server configuration and builds the tile grid and room graph.
func (mm *MapManager) generateWorldMap(configJSON string) *WorldMap {
	var config MapConfig
	_ = json.Unmarshal([]byte(configJSON), &config)

	for k, npcCfg := range config.NPCs {
		npcCfg.Key = k
		config.NPCs[k] = npcCfg
	}
	mm.NPCConfigs = config.NPCs
	mm.ItemConfigs = config.Items
	mm.QuestConfigs = config.Quests

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

	for y := range mapHeight {
		grid[y] = make([]TileType, mapWidth)
		envIdx[y] = make([]int, mapWidth)
		for x := range mapWidth {
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
			Exits:       roomCfg.Exits,
			GridX:       coord[0],
			GridY:       coord[1],
			TileX:       tileX,
			TileY:       tileY,
			NPCs:        roomCfg.NPCs,
			Items:       roomCfg.Items,
		}
		worldMap.Rooms[roomID] = roomNode

		for ry := 0; ry < 4; ry++ {
			for rx := 0; rx < 4; rx++ {
				if tileY+ry >= 0 && tileY+ry < mapHeight && tileX+rx >= 0 && tileX+rx < mapWidth {
					grid[tileY+ry][tileX+rx] = TileRoom
				}
			}
		}
	}

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

// carvePath marks a horizontal-then-vertical route between two grid coordinates as path tiles.
// grid is the tile map; x1, y1 and x2, y2 are the start and end tile coordinates.
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

// loadGroundTextures loads the ground, path, and room textures used by the map.
func (mm *MapManager) loadGroundTextures() {
	for i := range 16 {
		key := fmt.Sprintf("%04b", i)
		path := mm.PathResolver.Resolve("ground", key+".png")
		mm.PathTextures[key] = mm.PathResolver.ResourceManager.Load(path)
	}
	mm.EmptyTexture = mm.PathResolver.ResourceManager.Load(mm.PathResolver.Resolve("ground", "lawn.png"))

	fullPath := mm.PathResolver.Resolve("ground", "full.png")
	mm.RoomTextures["full"] = mm.PathResolver.ResourceManager.Load(fullPath)
}

// loadEnvironmentObjects loads the forest scenery textures and their display scales.
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
	}

	mm.SceneryObjects = make([]EnvObjectInfo, len(forestFiles))
	for i, file := range forestFiles {
		fullPath := mm.PathResolver.Resolve("forest", file.FileName)
		mm.SceneryObjects[i] = EnvObjectInfo{
			Path:     fullPath,
			Category: CategoryScenery,
			Scale:    file.Scale,
			Texture:  mm.PathResolver.ResourceManager.Load(fullPath),
		}
	}
}
