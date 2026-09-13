package component

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
	Update(dt float32)
	Render()
	UnloadAnimations()
	SetAction(action string)
}

type NpcCharacter struct {
	Name         string
	Position     rl.Vector2
	Scale        float32
	Animations   map[string][]rl.Texture2D
	CurrentAnim  string
	FrameIndex   int
	FrameTimer   float32
	FrameSpeed   float32
	PathResolver *utils.PathResolver
	RelativeDir  string
}

func NewNpcCharacter(name string, resolver *utils.PathResolver, relativeDir string, x, y, scale float32) NpcCharacter {
	return NpcCharacter{
		Name:         name,
		Position:     rl.NewVector2(x, y),
		Scale:        scale,
		Animations:   make(map[string][]rl.Texture2D),
		CurrentAnim:  "Idle",
		FrameSpeed:   24.0, 
		PathResolver: resolver,
		RelativeDir:  relativeDir,
	}
}

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
		textures = append(textures, rl.LoadTexture(texPath))
	}

	n.Animations[action] = textures
	return nil
}

func (n *NpcCharacter) SetAction(action string) {
	if _, exists := n.Animations[action]; exists && n.CurrentAnim != action {
		n.CurrentAnim = action
		n.FrameIndex = 0
		n.FrameTimer = 0
	}
}

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

	
	label := fmt.Sprintf("%s\n[%s]", strings.ToUpper(n.Name), n.CurrentAnim)
	textWidth := rl.MeasureText(n.Name, 18)
	labelX := int32(n.Position.X + (destWidth / 2) - float32(textWidth/2))
	labelY := int32(n.Position.Y - 45)
	rl.DrawText(label, labelX, labelY, 18, rl.DarkGray)
}

func (n *NpcCharacter) UnloadAnimations() {
	for _, frames := range n.Animations {
		for _, tex := range frames {
			rl.UnloadTexture(tex)
		}
	}
}