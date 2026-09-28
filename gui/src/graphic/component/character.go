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
	Position       rl.Vector2
	PositionOffset rl.Vector2
	Scale          float32
	Speed          float32
	PathResolver   *utils.PathResolver
	Direction      Direction
	Action         Action
	Animations     map[string][]rl.Texture2D
	FrameIndex     int
	FrameTimer     float32
	FrameSpeed     float32
	IsDead         bool
}

func NewCharacter(scale float32, resolver *utils.PathResolver, assetPath string) Character {
	if resolver == nil {
		resolver = utils.NewPathResolver(assetPath)
	} else {
		resolver = utils.NewPathResolverWithResourceManager(assetPath, resolver.ResourceManager)
	}
	return Character{
		Scale:        scale,
		Speed:        200.0,
		PathResolver: resolver,
		Direction:    DirFront,
		Action:       ActionIdle,
		Animations:   make(map[string][]rl.Texture2D),
		FrameSpeed:   80.0,
	}
}

func (c *Character) SetPosition(x, y float32) {
	c.Position = rl.NewVector2(x+c.PositionOffset.X, y+c.PositionOffset.Y)
}

func (c *Character) SetPositionOffset(x, y float32) {
	c.PositionOffset = rl.NewVector2(x, y)
}

func (c *Character) LoadAnim(key, subFolder, filePattern string, count int) {
	frames := make([]rl.Texture2D, count)
	for i := range count {
		path := c.PathResolver.ResolveFramePath(subFolder, filePattern, i)
		frames[i] = c.PathResolver.ResourceManager.Load(path)
	}
	c.Animations[key] = frames
}

func (c *Character) UnloadAnimations() {
	c.Animations = make(map[string][]rl.Texture2D)
}

func (c *Character) SetAction(newAction Action) {
	if c.Action != newAction {
		c.Action = newAction
		c.FrameIndex = 0
		c.FrameTimer = 0
	}
}
