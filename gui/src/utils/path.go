// Package utils provides shared resource, path, and theme helpers for the client.
package utils

import (
	"fmt"
	"path/filepath"
)

type PathResolver struct {
	BaseDir         string
	ResourceManager *ResourceManager
}

// NewPathResolver creates a path resolver rooted at baseDir with a new resource manager.
// baseDir is the root directory used to resolve asset paths.
func NewPathResolver(baseDir string) *PathResolver {
	return NewPathResolverWithResourceManager(baseDir, NewResourceManager())
}

// NewPathResolverWithResourceManager creates a resolver using baseDir and resourceManager.
// baseDir is the root asset directory; resourceManager loads and caches textures.
// If resourceManager is nil, a new resource manager is created.
func NewPathResolverWithResourceManager(baseDir string, resourceManager *ResourceManager) *PathResolver {
	if resourceManager == nil {
		resourceManager = NewResourceManager()
	}
	return &PathResolver{
		BaseDir:         baseDir,
		ResourceManager: resourceManager,
	}
}

// Resolve joins the resolver's base directory with the supplied path components.
// subPaths are path elements appended to the base directory.
func (p *PathResolver) Resolve(subPaths ...string) string {
	paths := append([]string{p.BaseDir}, subPaths...)
	return filepath.Join(paths...)
}

// ResolveFramePath builds the path for an indexed animation frame.
// subFolder is the animation directory; filePattern formats the frame filename;
// index is the frame number passed to filePattern.
func (p *PathResolver) ResolveFramePath(subFolder, filePattern string, index int) string {
	fileName := fmt.Sprintf(filePattern, index)
	return filepath.Join(p.BaseDir, subFolder, fileName)
}
