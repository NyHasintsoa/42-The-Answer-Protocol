package character

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"tap-gui/src/utils"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type INpc interface {
	// Update advances the active animation by dt seconds.
	Update(dt float32)
	// Render draws the character at its current position.
	Render()
	// UnloadAnimations clears the character's animation frames.
	UnloadAnimations()
	// SetAction selects the animation identified by action.
	SetAction(action string)
	// SetPosition places the character at x and y, including its position offset.
	SetPosition(x, y float32)
	// SetPositionOffset sets the horizontal and vertical position offsets.
	SetPositionOffset(x, y float32)
}

type NpcCharacter struct {
	Name           string
	Position       rl.Vector2
	PositionOffset rl.Vector2
	Scale          float32
	Animations     map[string][]rl.Texture2D
	CurrentAnim    string
	FrameIndex     int
	FrameTimer     float32
	FrameSpeed     float32
	PathResolver   *utils.PathResolver
	RelativeDir    string
}

// NewNpcCharacter creates an NPC character using the supplied name, resolver, and asset directory.
// name is the display name; resolver locates and loads assets; relativeDir identifies the NPC's
// asset directory relative to the resolver's base; scale controls rendered size.
func NewNpcCharacter(name string, resolver *utils.PathResolver, relativeDir string, scale float32) NpcCharacter {
	return NpcCharacter{
		Name:         name,
		Position:     rl.NewVector2(0, 0),
		Scale:        scale,
		Animations:   make(map[string][]rl.Texture2D),
		CurrentAnim:  "Idle",
		FrameSpeed:   24.0,
		PathResolver: resolver,
		RelativeDir:  relativeDir,
	}
}

// LoadAnim loads the PNG frames for action from the character's asset directory.
// action is both the animation key and the animation subdirectory.
func (n *NpcCharacter) LoadAnim(action string) error {
	dirPath := n.PathResolver.Resolve(n.RelativeDir, action)
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory '%s': %w", dirPath, err)
	}

	var fileNames []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".png") {
			fileNames = append(fileNames, entry.Name())
		}
	}
	sort.Strings(fileNames)

	var textures []rl.Texture2D
	for _, name := range fileNames {
		texPath := filepath.Join(dirPath, name)
		textures = append(textures, n.PathResolver.ResourceManager.Load(texPath))
	}

	n.Animations[action] = textures
	return nil
}

// SetAction selects action when animation frames for it have been loaded.
func (n *NpcCharacter) SetAction(action string) {
	if _, exists := n.Animations[action]; exists && n.CurrentAnim != action {
		n.CurrentAnim = action
		n.FrameIndex = 0
		n.FrameTimer = 0
	}
}

// SetPosition places the character at the specified coordinates, including its offset.
// x and y are the horizontal and vertical coordinates.
func (n *NpcCharacter) SetPosition(x, y float32) {
	n.Position = rl.NewVector2(x+n.PositionOffset.X, y+n.PositionOffset.Y)
}

// SetPositionOffset sets the horizontal and vertical position offsets.
// x and y are the offset values applied by SetPosition.
func (n *NpcCharacter) SetPositionOffset(x, y float32) {
	n.PositionOffset = rl.NewVector2(x, y)
}

// Update advances the active animation by dt seconds.
func (n *NpcCharacter) Update(dt float32) {
	frames := n.Animations[n.CurrentAnim]
	if len(frames) == 0 {
		return
	}

	n.FrameTimer += dt
	if n.FrameTimer >= (1.0 / n.FrameSpeed) {
		n.FrameTimer = 0
		n.FrameIndex = (n.FrameIndex + 1) % len(frames)
	}
}

// Render draws the current animation frame at the character's position.
func (n *NpcCharacter) Render() {
	frames, exists := n.Animations[n.CurrentAnim]
	if !exists || len(frames) == 0 {
		return
	}

	tex := frames[n.FrameIndex]

	srcRec := rl.NewRectangle(0, 0, float32(tex.Width), float32(tex.Height))

	destWidth := float32(tex.Width) * n.Scale
	destHeight := float32(tex.Height) * n.Scale

	destRec := rl.NewRectangle(n.Position.X, n.Position.Y, destWidth, destHeight)
	origin := rl.NewVector2(0, 0)

	rl.DrawTexturePro(tex, srcRec, destRec, origin, 0.0, rl.White)
}

// UnloadAnimations clears the character's animation frame references.
func (n *NpcCharacter) UnloadAnimations() {
	n.Animations = make(map[string][]rl.Texture2D)
}
