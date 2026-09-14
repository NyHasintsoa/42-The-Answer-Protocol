package utils

import (
	"path/filepath"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type ImageManager struct {
	textures map[string]rl.Texture2D
}

func NewImageManager() *ImageManager {
	return &ImageManager{textures: make(map[string]rl.Texture2D)}
}

func (im *ImageManager) Load(path string) rl.Texture2D {
	key := filepath.Clean(path)
	if texture, exists := im.textures[key]; exists {
		return texture
	}

	texture := rl.LoadTexture(path)
	im.textures[key] = texture
	return texture
}

func (im *ImageManager) UnloadAll() {
	for _, texture := range im.textures {
		if texture.ID > 0 {
			rl.UnloadTexture(texture)
		}
	}
	im.textures = make(map[string]rl.Texture2D)
}
