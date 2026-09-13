package component

import (
	"strings"

	"tap-gui/src/service"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type MapComponent struct {
	Manager *service.MapManager
}

func NewMapComponent(mgr *service.MapManager) *MapComponent {
	return &MapComponent{
		Manager: mgr,
	}
}

func (mc *MapComponent) Render() {
	if mc.Manager == nil || mc.Manager.World == nil {
		return
	}

	mm := mc.Manager

	
	for y := 0; y < mm.World.Height; y++ {
		for x := 0; x < mm.World.Width; x++ {
			pos := rl.NewVector2(float32(x)*service.TileSize, float32(y)*service.TileSize)
			destRect := rl.NewRectangle(pos.X, pos.Y, service.TileSize, service.TileSize)
			tile := mm.World.Grid[y][x]

			switch tile {
			case service.TilePath:
				mask := mm.GetPathBitmaskWorld(x, y)
				if tex, exists := mm.PathTextures[mask]; exists && tex.ID > 0 {
					srcRect := rl.NewRectangle(0, 0, float32(tex.Width), float32(tex.Height))
					rl.DrawTexturePro(tex, srcRect, destRect, rl.Vector2Zero(), 0, rl.White)
				} else {
					rl.DrawRectangleRec(destRect, rl.NewColor(160, 120, 80, 255))
				}

			case service.TileRoom, service.TileWall:
				if mm.EmptyTexture.ID > 0 {
					srcEmpty := rl.NewRectangle(0, 0, float32(mm.EmptyTexture.Width), float32(mm.EmptyTexture.Height))
					rl.DrawTexturePro(mm.EmptyTexture, srcEmpty, destRect, rl.Vector2Zero(), 0, rl.White)
				} else {
					col := rl.NewColor(34, 139, 34, 255)
					if tile == service.TileRoom {
						col = rl.NewColor(200, 170, 120, 255)
					}
					rl.DrawRectangleRec(destRect, col)
				}

				if tile == service.TileWall {
					idx := mm.World.EnvObjectIdx[y][x]
					if idx >= 0 && idx < len(mm.SceneryObjects) {
						obj := mm.SceneryObjects[idx]
						if obj.Texture.ID > 0 {
							srcRect := rl.NewRectangle(0, 0, float32(obj.Texture.Width), float32(obj.Texture.Height))
							scale := (service.TileSize / float32(obj.Texture.Width)) * obj.Scale * 3.0
							drawW := float32(obj.Texture.Width) * scale
							drawH := float32(obj.Texture.Height) * scale
							drawRect := rl.NewRectangle(pos.X+service.TileSize/2-drawW/2, pos.Y+service.TileSize/2-drawH/2, drawW, drawH)
							rl.DrawTexturePro(obj.Texture, srcRect, drawRect, rl.Vector2Zero(), 0, rl.White)
						}
					}
				}
			}
		}
	}

	
	for _, room := range mm.World.Rooms {
		pos := rl.NewVector2(float32(room.TileX)*service.TileSize, float32(room.TileY)*service.TileSize)
		bldSize := service.TileSize * 4.0

		var tex rl.Texture2D
		nameLower := strings.ToLower(room.Name)
		if strings.Contains(nameLower, "shop") {
			tex = mm.BuildingTextures["shop.png"]
		} else if strings.Contains(nameLower, "tavern") {
			tex = mm.BuildingTextures["Tavern.png"]
		} else if strings.Contains(nameLower, "boss") {
			tex = mm.BuildingTextures["Castle-Round.png"]
		} else if strings.Contains(nameLower, "cave") {
			tex = mm.BuildingTextures["Tent.png"]
		} else {
			tex = mm.BuildingTextures["House.png"]
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

	
	for _, n := range mm.NPCs {
		n.Render()
	}
	for _, e := range mm.Enemies {
		e.Render()
	}
}