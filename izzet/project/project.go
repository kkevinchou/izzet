package project

import (
	"path/filepath"
	"strings"

	"github.com/kkevinchou/kitolib/modelspec"
)

type Content struct {
	Name        string
	OutFilepath string
	InFilePath  string

	PeripheralFiles []string
}

// type ContentBrowser struct {
// 	Content map[string]Content
// }

type Project struct {
	Content   []Content
	WorldFile string
	Name      string
	// ContentBrowser ContentBrowser
	// World          world.GameWorld
}

func NewProject() *Project {
	return &Project{Content: []Content{}}
}

func (p *Project) AddGLTFContent(sourceFile string, document *modelspec.Document) {
	baseFileName := strings.Split(filepath.Base(sourceFile), ".")[0]
	p.Content = append(p.Content,
		Content{
			Name:            baseFileName,
			InFilePath:      sourceFile,
			PeripheralFiles: document.PeripheralFiles,
		})
}
