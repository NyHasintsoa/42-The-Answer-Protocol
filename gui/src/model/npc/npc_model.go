package npc

import (
	"tap-gui/src/utils"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type NPCConfig struct {
	Key           string   `json:"key,omitempty"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Type          string   `json:"type"`
	ImageType     string   `json:"image_type"`
	CanBeAttacked bool     `json:"can_be_attacked"`
	HP            int      `json:"hp"`
	MaxHP         int      `json:"max_hp"`
	AttackDamage  int      `json:"attack_damage"`
	QuestID       string   `json:"quest_id"`
	Dialogue      []string `json:"dialogue"`
}

type Monster interface {
	UpdateAI(dt float32, target rl.Vector2, isWalkable func(float32, float32) bool)
	Render()
	UnloadAnimations()
	SetPosition(x, y float32)
}

type INpc interface {
	Update(dt float32)
	Render()
	UnloadAnimations()
	SetPosition(x, y float32)
}

type EntityFactory struct {
	NewMonster func(scale float32, resolver *utils.PathResolver, assetPath, name string) Monster
	NewNPC     func(kind, relativeDir string, scale float32) (INpc, error)
}
