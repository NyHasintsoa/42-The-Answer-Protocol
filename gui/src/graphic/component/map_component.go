package component

import (
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

			if mm.EmptyTexture.ID > 0 {
				srcEmpty := rl.NewRectangle(0, 0, float32(mm.EmptyTexture.Width), float32(mm.EmptyTexture.Height))
				rl.DrawTexturePro(mm.EmptyTexture, srcEmpty, destRect, rl.Vector2Zero(), 0, rl.White)
			} else {
				rl.DrawRectangleRec(destRect, rl.NewColor(34, 139, 34, 255))
			}

			switch tile {
			case service.TileRoom:
				if tex, ok := mm.GetRoomTileTexture(x, y); ok && tex.ID > 0 {
					srcRect := rl.NewRectangle(0, 0, float32(tex.Width), float32(tex.Height))
					rl.DrawTexturePro(tex, srcRect, destRect, rl.Vector2Zero(), 0, rl.White)
				}

			case service.TilePath:
				if tex, ok := mm.GetPathTileTexture(x, y); ok && tex.ID > 0 {
					srcRect := rl.NewRectangle(0, 0, float32(tex.Width), float32(tex.Height))
					rl.DrawTexturePro(tex, srcRect, destRect, rl.Vector2Zero(), 0, rl.White)
				} else {
					rl.DrawRectangleRec(destRect, rl.NewColor(160, 120, 80, 255))
				}

			case service.TileWall:
				idx := mm.World.EnvObjectIdx[y][x]
				if idx >= 0 && idx < len(mm.SceneryObjects) {
					obj := mm.SceneryObjects[idx]
					if obj.Texture.ID > 0 {
						srcRect := rl.NewRectangle(0, 0, float32(obj.Texture.Width), float32(obj.Texture.Height))
						scale := (service.TileSize / float32(obj.Texture.Width)) * obj.Scale * 2.5
						drawW := float32(obj.Texture.Width) * scale
						drawH := float32(obj.Texture.Height) * scale
						drawRect := rl.NewRectangle(pos.X+service.TileSize/2-drawW/2, pos.Y+service.TileSize/2-drawH/2, drawW, drawH)
						rl.DrawTexturePro(obj.Texture, srcRect, drawRect, rl.Vector2Zero(), 0, rl.White)
					}
				}
			}
		}
	}

	for _, room := range mm.World.Rooms {
		roomRect := rl.NewRectangle(
			float32(room.TileX)*service.TileSize,
			float32(room.TileY)*service.TileSize,
			service.TileSize*4,
			service.TileSize*4,
		)

		isActive := (mm.ActiveRoom != nil && mm.ActiveRoom.ID == room.ID)

		if isActive {
			rl.DrawRectangleLinesEx(roomRect, 3, rl.Gold)
		}
		centerX := roomRect.X + roomRect.Width/2.0
		centerY := roomRect.Y + roomRect.Height/2.0

		nameSize := float32(rl.MeasureText(room.Name, 14))
		labelX := int32(centerX - nameSize/2)
		labelY := int32(centerY - 8)

		rl.DrawRectangle(labelX-4, labelY-2, int32(nameSize+8), 18, rl.NewColor(0, 0, 0, 180))
		rl.DrawText(room.Name, labelX, labelY, 14, rl.RayWhite)
	}

	for _, n := range mm.NPCs {
		n.Render()
	}
	for _, e := range mm.Enemies {
		e.Render()
	}
}