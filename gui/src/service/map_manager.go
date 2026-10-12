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
	NPCs map[string]npc.NPCConfig `json:"npcs"`
}

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

type WorldMap struct {
	Width        int
	Height       int
	Grid         [][]TileType
	Rooms        map[string]*RoomNode
	EnvObjectIdx [][]int
}

type MapManager struct {
	World          *WorldMap
	NPCConfigs     map[string]npc.NPCConfig
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

func (mm *MapManager) GetNPCConfig(key string) (npc.NPCConfig, bool) {
	if cfg, exists := mm.NPCConfigs[key]; exists {
		if cfg.Key == "" {
			cfg.Key = key
		}
		return cfg, true
	}
	return npc.NPCConfig{Key: key, Name: key}, false
}

func (mm *MapManager) DamageNPC(key string, damage int) (npc.NPCConfig, bool) {
	key = strings.ToLower(key)
	npc, exists := mm.NPCConfigs[key]
	if !exists || !npc.CanBeAttacked || damage <= 0 || (npc.MaxHP > 0 && npc.HP <= 0) {
		return npc, false
	}
	if npc.MaxHP > 0 {
		npc.HP -= damage
		if npc.HP < 0 {
			npc.HP = 0
		}
		if npc.HP == 0 {
			npc.CanBeAttacked = false
		}
	}
	mm.NPCConfigs[key] = npc
	return npc, true
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

func (mm *MapManager) generateWorldMap(configJSON string) *WorldMap {
	var config MapConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		_ = json.Unmarshal([]byte(defaultMapConfigJSON), &config)
	}

	for k, npcCfg := range config.NPCs {
		npcCfg.Key = k
		config.NPCs[k] = npcCfg
	}
	mm.NPCConfigs = config.NPCs

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
		mm.PathTextures[key] = mm.PathResolver.ResourceManager.Load(path)
	}
	mm.EmptyTexture = mm.PathResolver.ResourceManager.Load(mm.PathResolver.Resolve("ground", "lawn.png"))

	fullPath := mm.PathResolver.Resolve("ground", "full.png")
	mm.RoomTextures["full"] = mm.PathResolver.ResourceManager.Load(fullPath)
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
                "name": "Village Square",
                "description": "A bustling square with cobblestone paths, surrounded by old stone buildings and a suspicious potion puddle.",
                "exits": {
                    "north": "alchemist_shop",
                    "east": "shop",
                    "south": "forest_edge",
                    "west": "hidden_grove"
                },
                "items": ["potion_health_PM"],
                "npcs": ["guard", "mayor"]
            },
            "alchemist_shop": {
                "name": "Alchemist Shop",
                "description": "A cozy, bubbling workshop smelling of sulfur, glowing herbs, and explosive trial potions.",
                "exits": {
                    "south": "start",
                    "east": "library"
                },
                "items": ["potion_health_GM", "potion_attack_PM"],
                "npcs": ["alchemist", "person"]
            },
            "library": {
                "name": "Ancient Library",
                "description": "Dusty bookshelves stretch to the ceiling, filled with ancient combat manuals and coffee-stained scrolls.",
                "exits": {
                    "west": "alchemist_shop",
                    "south": "shop",
                    "north": "training_grounds"
                },
                "items": ["book_dagger", "book_sword"],
                "npcs": ["scholar"]
            },
            "training_grounds": {
                "name": "Training Grounds",
                "description": "A muddy yard filled with wooden dummies, targets full of arrows, and discarded practice weapons.",
                "exits": {
                    "south": "library"
                },
                "items": ["dagger"],
                "npcs": ["veteran"]
            },
            "shop": {
                "name": "General Store",
                "description": "A cluttered shop filled with oddities, heavy armor, and sharp iron weapons.",
                "exits": {
                    "west": "start",
                    "north": "library",
                    "south": "deep_forest"
                },
                "items": ["sword", "potion_attack_GM"],
                "npcs": ["merchant"]
            },
            "forest_edge": {
                "name": "Forest Edge",
                "description": "The trees start to thicken here. Strange giggling noises echo from deeper in the woods.",
                "exits": {
                    "north": "start",
                    "east": "deep_forest"
                },
                "items": [],
                "npcs": ["gobelin"]
            },
            "hidden_grove": {
                "name": "Hidden Grove",
                "description": "A quiet, misty grove where rare medicinal plants glow under the moonlight.",
                "exits": {
                    "east": "start"
                },
                "items": ["herb"],
                "npcs": ["hermit"]
            },
            "deep_forest": {
                "name": "Deep Forest",
                "description": "The woods are dark and full of shadows. You feel like several pairs of eyes are judging your outfit.",
                "exits": {
                    "north": "shop",
                    "west": "forest_edge",
                    "south": "cave_entrance"
                },
                "items": [],
                "npcs": ["gobelin_chief"]
            },
            "cave_entrance": {
                "name": "Cave Entrance",
                "description": "A gaping dark hole in the rock face emitting scorched heat and dragon snorts.",
                "exits": {
                    "north": "deep_forest",
                    "south": "boss_lair"
                },
                "items": [],
                "npcs": []
            },
            "boss_lair": {
                "name": "Boss Lair",
                "description": "A vast cavern filled with charred bones, shiny gold coins, and an extremely grumpy dragon.",
                "exits": {
                    "north": "cave_entrance"
                },
                "items": ["magic_amulet", "boss_horn"],
                "npcs": ["boss_dragon"]
            }
        }
    },
    "npcs": {
        "guard": {
            "name": "Village Guard",
            "description": "A stern-looking guard leaning heavily on a spear.",
            "type": "quest_giver",
            "image_type": "guard",
            "can_be_attacked": false,
            "quest_id": "fetch_herbs",
            "dialogue": [
                "I used to be an adventurer like you, but then I realized guard duty has a dental plan.",
                "Stand clear! I am guarding this cobblestone with my life. It is very special cobblestone.",
                "If you bring me that Healing Herb from the Hidden Grove, I'll reward you with a Petite Health Potion!"
            ]
        },
        "alchemist": {
            "name": "Mad Alchemist",
            "description": "A eccentric brewer wearing singed goggles and mixing glowing liquids.",
            "type": "merchant",
            "image_type": "merchant",
            "can_be_attacked": false,
            "dialogue": [
                "Welcome to my lab! Drink the green potion for health, and the red potion for attack power! Don't mix them unless you want extra arms.",
                "Petite Potions (PM) are great for quick boosts, but Grand Potions (GM) pack a real punch!",
                "No refunds if a potion turns your eyebrows blue for three weeks."
            ]
        },
        "merchant": {
            "name": "Shop Keeper",
            "description": "A friendly but suspiciously wealthy merchant.",
            "type": "merchant",
            "image_type": "merchant",
            "can_be_attacked": false,
            "dialogue": [
                "Welcome! Buy something or please stop breathing on my shiny weapons.",
                "Need weapons? Grab an Iron Sword or a Grand Attack Potion (GM) before venturing out!",
                "I accept gold, silver, or juicy gossip about the Mayor."
            ]
        },
        "mayor": {
            "name": "Village Mayor",
            "description": "The pompous mayor wearing a sash that is slightly too tight.",
            "type": "quest_giver",
            "image_type": "person",
            "can_be_attacked": false,
            "quest_id": "fetch_amulet",
            "dialogue": [
                "As Mayor, I decree that someone else should go into the dangerous cave and save us!",
                "My campaign promise was zero dragon attacks. Technically the dragon stays in its cave, so I am winning!",
                "Retrieve my Magic Amulet from the Cave Dragon and I shall award you official Village Medals of Honor!"
            ]
        },
        "scholar": {
            "name": "Royal Scholar",
            "description": "An old wizard in thick spectacles surrounded by piles of books.",
            "type": "quest_giver",
            "image_type": "person",
            "can_be_attacked": false,
            "quest_id": "scholar_research",
            "dialogue": [
                "Knowledge is power! Have you read 'Dagger Techniques' or 'Mastering the Blade' yet?",
                "If you can bring me the Dragon Horn from the Boss Lair, I will teach you secret weapon techniques!"
            ]
        },
        "veteran": {
            "name": "Master at Arms",
            "description": "A scarred warrior demonstrating sword maneuvers.",
            "type": "dialogue",
            "image_type": "guard",
            "can_be_attacked": false,
            "dialogue": [
                "A dagger in the dark is worth two swords in the scabbard!",
                "Always drink an Attack Potion before charging into a dragon's lair kid!"
            ]
        },
        "person": {
            "name": "Tipsy Bard",
            "description": "A local bard slumping near the alchemist's brewing vats.",
            "type": "dialogue",
            "image_type": "person",
            "can_be_attacked": false,
            "dialogue": [
                "*Hic* I drank an Attack Potion (PM) thinking it was apple juice, and now I punched a wooden fence down!",
                "Did you know if you chug a Grand Health Potion (GM), you feel like you could wrestle a gobelin with one hand?"
            ]
        },
        "hermit": {
            "name": "Misty Hermit",
            "description": "An old gardener picking glowing herbs in the mist.",
            "type": "dialogue",
            "image_type": "person",
            "can_be_attacked": false,
            "dialogue": [
                "The herbs here make the finest Health Potions in the realm. The guard loves them for his sore joints!"
            ]
        },
        "gobelin": {
            "name": "Nasty Gobelin",
            "description": "A twitchy green creature clutching a rusty dagger.",
            "type": "gobelin",
            "image_type": "gobelin",
            "can_be_attacked": true,
            "max_hp": 30,
            "attack_damage": 5,
            "dialogue": [
                "Grah! Give me your shiny potions or I will bite your ankles!",
                "You think you tough because you tall? Short gobelins strike knee level first!",
                "Hey! That's my shiny dagger! Drop it or face my pointy stick!"
            ]
        },
        "gobelin_chief": {
            "name": "Gobelin Chief",
            "description": "A bulkier gobelin wearing a helmet made from an old cooking pot.",
            "type": "gobelin",
            "image_type": "gobelin",
            "can_be_attacked": true,
            "max_hp": 55,
            "attack_damage": 10,
            "dialogue": [
                "I am Chief of Forest! No tall humans allowed without paying potion tax!"
            ]
        },
        "boss_dragon": {
            "name": "Cave Dragon",
            "description": "A massive, scaly beast sleeping atop a pile of stolen shiny trinkets.",
            "type": "boss",
            "image_type": "boss",
            "can_be_attacked": true,
            "max_hp": 150,
            "attack_damage": 20,
            "dialogue": [
                "WHO DARES DISTURB MY NAP?! Do you have any idea how hard it is to sleep on pointy amulets and dragon horns?!",
                "ROAR! I will roast you like a marshmallow on a stick!",
                "Another 'hero'? Great, more crunchy armor to pick out of my teeth..."
            ]
        }
    },
    "items": {
        "book_dagger": {
            "id": "book_dagger",
            "name": "Dagger Mastery Book",
            "description": "A leather-bound manual detailing quick strikes and stealth assassination maneuvers."
        },
        "book_sword": {
            "id": "book_sword",
            "name": "Sword Combat Manual",
            "description": "An ancient tome describing heavy swings, parries, and knightly stances."
        },
        "boss_horn": {
            "id": "boss_horn",
            "name": "Dragon Horn",
            "description": "A razor-sharp horn severed from the Cave Dragon. It radiates intense heat."
        },
        "dagger": {
            "id": "dagger",
            "name": "Rogue Dagger",
            "description": "A swift, light blade perfect for sneaky critical strikes."
        },
        "medals": {
            "id": "medals",
            "name": "Mayor's Medal of Honor",
            "description": "A shiny gold medal awarded by the Mayor. It looks expensive, but is mostly painted cardboard."
        },
        "potion_attack_PM": {
            "id": "potion_attack_PM",
            "name": "Petite Attack Potion",
            "description": "A small vial of glowing red elixir that temporarily boosts your attack damage (+5 ATK)."
        },
        "potion_attack_GM": {
            "id": "potion_attack_GM",
            "name": "Grand Attack Potion",
            "description": "A large flask of surging crimson brew that drastically empowers your strikes (+20 ATK)."
        },
        "potion_health_PM": {
            "id": "potion_health_PM",
            "name": "Petite Health Potion",
            "description": "A small green potion that instantly restores 25 HP."
        },
        "potion_health_GM": {
            "id": "potion_health_GM",
            "name": "Grand Health Potion",
            "description": "A potent, bubbling green brew that fully restores up to 100 HP."
        },
        "sword": {
            "id": "sword",
            "name": "Iron Sword",
            "description": "A sturdy iron sword. Far better than throwing harsh words at a dragon."
        },
        "herb": {
            "id": "herb",
            "name": "Healing Herb",
            "description": "A glowing medicinal plant emitting a soothing peppermint aroma."
        },
        "magic_amulet": {
            "id": "magic_amulet",
            "name": "Magic Amulet",
            "description": "A glowing ancient amulet belonging to the Mayor. It hums with tacky magical energy."
        }
    }
}`
