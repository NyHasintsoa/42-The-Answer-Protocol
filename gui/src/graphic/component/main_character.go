package component

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type MainCharacter struct {
	Character
}

func NewMainCharacter(x, y float32, scale float32, assetPath string) *MainCharacter {
	mc := &MainCharacter{
		Character: NewCharacter(x, y, scale, assetPath),
	}
	mc.loadAnimations()
	return mc
}

func (mc *MainCharacter) loadAnimations() {
	mc.LoadAnim("Front_Idle", "Front_Idle", "Front_Idle_%03d.png", 16)
	mc.LoadAnim("Back_Idle", "Back_Idle", "Back_Idle_%03d.png", 16)
	mc.LoadAnim("Left_Idle", "Left_Idle", "Left_Idle_%03d.png", 16)
	mc.LoadAnim("Right_Idle", "Right_Idle", "Right_Idle_%03d.png", 16)

	mc.LoadAnim("Front_Walking", "Front_Walking", "Front_Walking_%03d.png", 20)
	mc.LoadAnim("Back_Walking", "Back_Walking", "Back_Walking_%03d.png", 20)
	mc.LoadAnim("Left_Walking", "Left_Walking", "Left_Walking_%03d.png", 20)
	mc.LoadAnim("Right_Walking", "Right_Walking", "Right_Walking_%03d.png", 20)

	mc.LoadAnim("Front_Attacking", "Front_Attacking", "Front_Attacking_%03d.png", 10)
	mc.LoadAnim("Back_Attacking", "Back_Attacking", "Back_Attacking_%03d.png", 10)
	mc.LoadAnim("Left_Attacking", "Left_Attacking", "Left_Attacking_%03d.png", 10)
	mc.LoadAnim("Right_Attacking", "Right_Attacking", "Right_Attacking_%03d.png", 10)

	mc.LoadAnim("Front_Hurt", "Front_Hurt", "FrontHurt_%03d.png", 10)
	mc.LoadAnim("Back_Hurt", "Back_Hurt", "Back_Hurt_%03d.png", 10)
	mc.LoadAnim("Left_Hurt", "Left_Hurt", "Left_Hurt_%03d.png", 10)
	mc.LoadAnim("Right_Hurt", "Right_Hurt", "Right_Hurt_%03d.png", 10)

	mc.LoadAnim("Dying", "Dying", "Dying_%03d.png", 10)
}

func (mc *MainCharacter) Render() {
	key := mc.GetAnimKey()
	frames := mc.Animations[key]
	if len(frames) == 0 {
		return
	}

	if mc.FrameIndex >= len(frames) {
		mc.FrameIndex = 0
	}

	tex := frames[mc.FrameIndex]

	scaledWidth := float32(tex.Width) * mc.Scale
	scaledHeight := float32(tex.Height) * mc.Scale

	srcRect := rl.NewRectangle(0, 0, float32(tex.Width), float32(tex.Height))
	destRect := rl.NewRectangle(
		mc.Position.X,
		mc.Position.Y,
		scaledWidth,
		scaledHeight,
	)

	origin := rl.NewVector2(scaledWidth/2, scaledHeight/2+30)
	rl.DrawTexturePro(tex, srcRect, destRect, origin, 0, rl.White)
}

func (c *MainCharacter) GetAnimKey() string {
	if c.Action == ActionDying {
		return "Dying"
	}

	dirStr := ""
	switch c.Direction {
	case DirFront:
		dirStr = "Front"
	case DirBack:
		dirStr = "Back"
	case DirLeft:
		dirStr = "Left"
	case DirRight:
		dirStr = "Right"
	}

	actionStr := ""
	switch c.Action {
	case ActionIdle:
		actionStr = "Idle"
	case ActionWalking:
		actionStr = "Walking"
	case ActionAttacking:
		actionStr = "Attacking"
	case ActionHurt:
		actionStr = "Hurt"
	}

	return fmt.Sprintf("%s_%s", dirStr, actionStr)
}

func (c *MainCharacter) AdvanceFrame(dt float32, loop bool) bool {
	c.FrameTimer += dt
	if c.FrameTimer >= (1.0 / c.FrameSpeed) {
		c.FrameTimer = 0
		c.FrameIndex++
		animKey := c.GetAnimKey()
		frames := c.Animations[animKey]
		if c.FrameIndex >= len(frames) {
			if loop {
				c.FrameIndex = 0
			} else {
				c.FrameIndex = len(frames) - 1
				return true
			}
		}
	}
	return false
}