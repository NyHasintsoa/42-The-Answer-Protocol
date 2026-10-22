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

// NewResourceManager creates a resource manager with an empty texture cache.
func NewResourceManager() *ResourceManager {
	return &ResourceManager{textures: make(map[string]rl.Texture2D)}
}

// Load returns the cached texture for path or loads and caches it from disk.
// path is the image file to load.
func (rm *ResourceManager) Load(path string) rl.Texture2D {
	if rm == nil {
		return rl.Texture2D{}
	}
	if rm.textures == nil {
		rm.textures = make(map[string]rl.Texture2D)
	}
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

// PreloadDirectory loads every supported image under root before returning.
// root is the directory tree to scan; loading stops and returns an error on failure.
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

// PreloadNextBatch loads up to batchSize images from root and reports whether preloading is complete.
// root is the directory tree to scan; batchSize is the maximum number of images to load this call.
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

// PreloadProgress reports the number of images loaded and discovered.
func (rm *ResourceManager) PreloadProgress() (loaded, total int) {
	return rm.preloadIndex, len(rm.preloadPaths)
}

// isImageResource reports whether path has a supported image extension.
// path is the filesystem path to inspect.
func isImageResource(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpeg", ".jpg", ".png":
		return true
	default:
		return false
	}
}

// UnloadAll releases cached textures and resets the preload state.
func (rm *ResourceManager) UnloadAll() {
	if rm == nil {
		return
	}
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
