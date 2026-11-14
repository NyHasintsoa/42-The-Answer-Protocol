// Package character loads, animates, and renders player and non-player characters.
package character

import (
	"tap-gui/src/utils"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Action identifies the animation currently played by a character.
type Action int

// Direction identifies the direction a character faces.
type Direction int

const (
	// DirFront faces the character toward the viewer.
	DirFront Direction = iota
	// DirBack faces the character away from the viewer.
	DirBack
	// DirLeft faces the character left.
	DirLeft
	// DirRight faces the character right.
	DirRight
)

const (
	// ActionIdle plays the character's idle animation.
	ActionIdle Action = iota
	// ActionWalking plays the character's walking animation.
	ActionWalking
	// ActionAttacking plays the character's attack animation.
	ActionAttacking
	// ActionHurt plays the character's hurt animation.
	ActionHurt
	// ActionDying plays the character's death animation.
	ActionDying
)

// Character stores shared position, animation, and movement state.
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

// NewCharacter creates a character with the requested scale and asset directory.
// scale controls rendered size; resolver supplies resource lookup when available;
// assetPath is the directory used to resolve animation assets.
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

// SetPosition moves c to the specified coordinates, including its configured offset.
// x and y are the horizontal and vertical coordinates.
func (c *Character) SetPosition(x, y float32) {
	c.Position = rl.NewVector2(x+c.PositionOffset.X, y+c.PositionOffset.Y)
}

// SetPositionOffset sets the offset added to future positions.
// x and y are the horizontal and vertical offset values.
func (c *Character) SetPositionOffset(x, y float32) {
	c.PositionOffset = rl.NewVector2(x, y)
}

// LoadAnim loads count animation frames and stores them under key.
// key identifies the animation; subFolder is its asset directory; filePattern formats
// each frame filename using its zero-based index; count is the number of frames.
func (c *Character) LoadAnim(key, subFolder, filePattern string, count int) {
	frames := make([]rl.Texture2D, count)
	for i := range count {
		path := c.PathResolver.ResolveFramePath(subFolder, filePattern, i)
		frames[i] = c.PathResolver.ResourceManager.Load(path)
	}
	c.Animations[key] = frames
}

// UnloadAnimations clears the character's references to loaded animation frames.
func (c *Character) UnloadAnimations() {
	c.Animations = make(map[string][]rl.Texture2D)
}

// SetAction changes the current animation action and resets its frame when it changes.
// newAction is the action to display.
func (c *Character) SetAction(newAction Action) {
	if c.Action != newAction {
		c.Action = newAction
		c.FrameIndex = 0
		c.FrameTimer = 0
	}
}
