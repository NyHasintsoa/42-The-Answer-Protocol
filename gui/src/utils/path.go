package utils

import (
	"fmt"
	"path/filepath"
)

type PathResolver struct {
	BaseDir      string
	ImageManager *ImageManager
}

func NewPathResolver(baseDir string) *PathResolver {
	return NewPathResolverWithImageManager(baseDir, NewImageManager())
}

func NewPathResolverWithImageManager(baseDir string, imageManager *ImageManager) *PathResolver {
	if imageManager == nil {
		imageManager = NewImageManager()
	}
	return &PathResolver{
		BaseDir:      baseDir,
		ImageManager: imageManager,
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
