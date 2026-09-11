package utils

import (
	"fmt"
	"path/filepath"
)

type PathResolver struct {
	BaseDir string
}

func NewPathResolver(baseDir string) *PathResolver {
	return &PathResolver{
		BaseDir: baseDir,
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