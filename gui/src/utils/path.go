package utils

import (
	"fmt"
	"path/filepath"
)

type PathResolver struct {
	BaseDir         string
	ResourceManager *ResourceManager
}

func NewPathResolver(baseDir string) *PathResolver {
	return NewPathResolverWithResourceManager(baseDir, NewResourceManager())
}

func NewPathResolverWithResourceManager(baseDir string, resourceManager *ResourceManager) *PathResolver {
	if resourceManager == nil {
		resourceManager = NewResourceManager()
	}
	return &PathResolver{
		BaseDir:         baseDir,
		ResourceManager: resourceManager,
	}
}

func (p *PathResolver) Resolve(subPaths ...string) string {
	paths := append([]string{p.BaseDir}, subPaths...)
	return filepath.Join(paths...)
}

func (p *PathResolver) ResolveFramePath(subFolder, filePattern string, index int) string {
	fileName := fmt.Sprintf(filePattern, index)
	return filepath.Join(p.BaseDir, subFolder, fileName)
}
