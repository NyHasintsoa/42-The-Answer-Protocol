package component

import (
	"tap-gui/src/utils"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Action int
type Direction int

const (
	DirFront Direction = iota
	DirBack
	DirLeft
	DirRight
)

const (
	ActionIdle Action = iota
	ActionWalking
	ActionAttacking
	ActionHurt
	ActionDying
)

type Character struct {
	Position     rl.Vector2
	Scale        float32
	Speed        float32
	PathResolver *utils.PathResolver
	Direction    Direction
	Action       Action
	Animations   map[string][]rl.Texture2D
	FrameIndex   int
	FrameTimer   float32
	FrameSpeed   float32
	IsDead       bool
}

func NewCharacter(x, y float32, scale float32, assetPath string) Character {
	return Character{
		Position:     rl.NewVector2(x, y),
		Scale:        scale,
		Speed:        200.0,
		PathResolver: utils.NewPathResolver(assetPath),
		Direction:    DirFront,
		Action:       ActionIdle,
		Animations:   make(map[string][]rl.Texture2D),
		FrameSpeed:   80.0,
	}
}

func (c *Character) LoadAnim(key, subFolder, filePattern string, count int) {
	frames := make([]rl.Texture2D, count)
	for i := range count {
		path := c.PathResolver.ResolveFramePath(subFolder, filePattern, i)
		frames[i] = rl.LoadTexture(path)
	}
	c.Animations[key] = frames
}

func (c *Character) UnloadAnimations() {
	for _, frames := range c.Animations {
		for _, tex := range frames {
			rl.UnloadTexture(tex)
		}
	}
}

func (c *Character) SetAction(newAction Action) {
	if c.Action != newAction {
		c.Action = newAction
		c.FrameIndex = 0
		c.FrameTimer = 0
	}
}