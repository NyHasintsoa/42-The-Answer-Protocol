package utils

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type ResourceManager struct {
	textures     map[string]rl.Texture2D
	preloadPaths []string
	preloadIndex int
	preloadReady bool
}

func NewResourceManager() *ResourceManager {
	return &ResourceManager{textures: make(map[string]rl.Texture2D)}
}

func (rm *ResourceManager) Load(path string) rl.Texture2D {
	key := filepath.Clean(path)
	if texture, exists := rm.textures[key]; exists {
		return texture
	}

	texture := rl.LoadTexture(path)
	if texture.ID > 0 {
		rm.textures[key] = texture
	}
	return texture
}

func (rm *ResourceManager) PreloadDirectory(root string) error {
	for {
		done, err := rm.PreloadNextBatch(root, 64)
		if err != nil {
			return err
		}
		if done {
			return nil
		}
	}
}

func (rm *ResourceManager) PreloadNextBatch(root string, batchSize int) (bool, error) {
	if !rm.preloadReady {
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !entry.IsDir() && isImageResource(path) {
				rm.preloadPaths = append(rm.preloadPaths, path)
			}
			return nil
		})
		if err != nil {
			return false, err
		}
		rm.preloadReady = true
	}

	if batchSize < 1 {
		batchSize = 1
	}
	for loaded := 0; loaded < batchSize && rm.preloadIndex < len(rm.preloadPaths); loaded++ {
		path := rm.preloadPaths[rm.preloadIndex]
		if texture := rm.Load(path); texture.ID == 0 {
			return false, fmt.Errorf("failed to load image resource %q", path)
		}
		rm.preloadIndex++
	}
	return rm.preloadIndex == len(rm.preloadPaths), nil
}

func (rm *ResourceManager) PreloadProgress() (loaded, total int) {
	return rm.preloadIndex, len(rm.preloadPaths)
}

func isImageResource(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpeg", ".jpg", ".png":
		return true
	default:
		return false
	}
}

func (rm *ResourceManager) UnloadAll() {
	for _, texture := range rm.textures {
		if texture.ID > 0 {
			rl.UnloadTexture(texture)
		}
	}
	rm.textures = make(map[string]rl.Texture2D)
	rm.preloadPaths = nil
	rm.preloadIndex = 0
	rm.preloadReady = false
}
