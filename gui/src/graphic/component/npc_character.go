package component

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type NPC struct {
	Character
	Name string
}

func NewNPC(x, y float32, scale float32, assetPath, name string) *NPC {
	npc := &NPC{
		Character: NewCharacter(x, y, scale, assetPath),
		Name:      name,
	}
	npc.loadAnimations()
	return npc
}

func (npc *NPC) loadAnimations() {
	npc.LoadAnim("Front_Idle", "Front_Idle", "Front_Idle_%03d.png", 16)
	npc.LoadAnim("Back_Idle", "Back_Idle", "Back_Idle_%03d.png", 16)
	npc.LoadAnim("Left_Idle", "Left_Idle", "Left_Idle_%03d.png", 16)
	npc.LoadAnim("Right_Idle", "Right_Idle", "Right_Idle_%03d.png", 16)

	npc.LoadAnim("Front_Walking", "Front_Walking", "Front_Walking_%03d.png", 20)
	npc.LoadAnim("Back_Walking", "Back_Walking", "Back_Walking_%03d.png", 20)
	npc.LoadAnim("Left_Walking", "Left_Walking", "Left_Walking_%03d.png", 20)
	npc.LoadAnim("Right_Walking", "Right_Walking", "Right_Walking_%03d.png", 20)
}

func (npc *NPC) GetAnimKey() string {
	dirStr := ""
	switch npc.Direction {
	case DirFront:
		dirStr = "Front"
	case DirBack:
		dirStr = "Back"
	case DirLeft:
		dirStr = "Left"
	case DirRight:
		dirStr = "Right"
	}

	actionStr := "Idle"
	if npc.Action == ActionWalking {
		actionStr = "Walking"
	}

	return fmt.Sprintf("%s_%s", dirStr, actionStr)
}

func (npc *NPC) Draw() {
	key := npc.GetAnimKey()
	frames := npc.Animations[key]
	
	if len(frames) == 0 {
		rl.DrawRectangle(int32(npc.Position.X-20), int32(npc.Position.Y-20), 40, 40, rl.Blue)
		rl.DrawText(npc.Name, int32(npc.Position.X-20), int32(npc.Position.Y)-35, 12, rl.White)
		return
	}

	if npc.FrameIndex >= len(frames) {
		npc.FrameIndex = 0
	}

	tex := frames[npc.FrameIndex]
	scaledWidth := float32(tex.Width) * npc.Scale
	scaledHeight := float32(tex.Height) * npc.Scale

	srcRect := rl.NewRectangle(0, 0, float32(tex.Width), float32(tex.Height))
	destRect := rl.NewRectangle(
		npc.Position.X,
		npc.Position.Y,
		scaledWidth,
		scaledHeight,
	)

	origin := rl.NewVector2(scaledWidth/2, scaledHeight/2+30)
	rl.DrawTexturePro(tex, srcRect, destRect, origin, 0, rl.White)

	// Draw NPC name
	rl.DrawText(npc.Name, int32(npc.Position.X-float32(rl.MeasureText(npc.Name, 12))/2), int32(npc.Position.Y-scaledHeight/2-10), 12, rl.SkyBlue)
}

func (npc *NPC) AdvanceFrame(dt float32, loop bool) bool {
	npc.FrameTimer += dt
	if npc.FrameTimer >= (1.0 / npc.FrameSpeed) {
		npc.FrameTimer = 0
		npc.FrameIndex++
		frames := npc.Animations[npc.GetAnimKey()]
		if npc.FrameIndex >= len(frames) {
			if loop {
				npc.FrameIndex = 0
			} else {
				npc.FrameIndex = len(frames) - 1
				return true
			}
		}
	}
	return false
}