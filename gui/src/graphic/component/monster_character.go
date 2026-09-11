package component

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Monster struct {
	Character
	Name string
}

func NewMonster(x, y float32, scale float32, assetPath string, name string) *Monster {
	m := &Monster{
		Character: NewCharacter(x, y, scale, assetPath),
		Name:      name,
	}
	m.Speed = 120.0
	m.loadAnimations()
	return m
}

func (m *Monster) loadAnimations() {
	m.LoadAnim("Front_Idle", "Front_Idle", "Front_Idle_%03d.png", 16)
	m.LoadAnim("Back_Idle", "Back_Idle", "Back_Idle_%03d.png", 16)
	m.LoadAnim("Left_Idle", "Left_Idle", "Left_Idle_%03d.png", 16)
	m.LoadAnim("Right_Idle", "Right_Idle", "Right_Idle_%03d.png", 16)

	m.LoadAnim("Front_Walking", "Front_Walking", "Front_Walking_%03d.png", 20)
	m.LoadAnim("Back_Walking", "Back_Walking", "Back_Walking_%03d.png", 20)
	m.LoadAnim("Left_Walking", "Left_Walking", "Left_Walking_%03d.png", 20)
	m.LoadAnim("Right_Walking", "Right_Walking", "Right_Walking_%03d.png", 20)

	m.LoadAnim("Front_Attacking", "Front_Attacking", "Front_Attacking_%03d.png", 10)
	m.LoadAnim("Back_Attacking", "Back_Attacking", "Back_Attacking_%03d.png", 10)
	m.LoadAnim("Left_Attacking", "Left_Attacking", "Left_Attacking_%03d.png", 10)
	m.LoadAnim("Right_Attacking", "Right_Attacking", "Right_Attacking_%03d.png", 10)

	m.LoadAnim("Front_Hurt", "Front_Hurt", "Front_Hurt_%03d.png", 10)
	m.LoadAnim("Back_Hurt", "Back_Hurt", "Back_Hurt_%03d.png", 10)
	m.LoadAnim("Left_Hurt", "Left_Hurt", "Left_Hurt_%03d.png", 10)
	m.LoadAnim("Right_Hurt", "Right_Hurt", "Right_Hurt_%03d.png", 10)

	m.LoadAnim("Dying", "Dying", "Dying_%03d.png", 10)
}

func (m *Monster) GetAnimKey() string {
	if m.Action == ActionDying {
		return "Dying"
	}

	dirStr := ""
	switch m.Direction {
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
	switch m.Action {
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

func (m *Monster) Draw() {
	key := m.GetAnimKey()
	frames := m.Animations[key]
	
	if len(frames) == 0 {
		rl.DrawRectangle(int32(m.Position.X-20), int32(m.Position.Y-20), 40, 40, rl.Red)
		rl.DrawText(m.Name, int32(m.Position.X-20), int32(m.Position.Y)-35, 12, rl.White)
		return
	}

	if m.FrameIndex >= len(frames) {
		m.FrameIndex = 0
	}

	tex := frames[m.FrameIndex]
	scaledWidth := float32(tex.Width) * m.Scale
	scaledHeight := float32(tex.Height) * m.Scale

	srcRect := rl.NewRectangle(0, 0, float32(tex.Width), float32(tex.Height))
	destRect := rl.NewRectangle(
		m.Position.X,
		m.Position.Y,
		scaledWidth,
		scaledHeight,
	)

	origin := rl.NewVector2(scaledWidth/2, scaledHeight/2+30)
	rl.DrawTexturePro(tex, srcRect, destRect, origin, 0, rl.White)
	
	rl.DrawText(m.Name, int32(m.Position.X-float32(rl.MeasureText(m.Name, 12))/2), int32(m.Position.Y-scaledHeight/2-10), 12, rl.Red)
}

func (m *Monster) AdvanceFrame(dt float32, loop bool) bool {
	m.FrameTimer += dt
	if m.FrameTimer >= (1.0 / m.FrameSpeed) {
		m.FrameTimer = 0
		m.FrameIndex++
		frames := m.Animations[m.GetAnimKey()]
		if m.FrameIndex >= len(frames) {
			if loop {
				m.FrameIndex = 0
			} else {
				m.FrameIndex = len(frames) - 1
				return true
			}
		}
	}
	return false
}