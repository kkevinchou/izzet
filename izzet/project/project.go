package project

import (
	"path/filepath"
	"strings"
)

type Content struct {
	Name        string
	OutFilepath string
	InFilePath  string
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

func (p *Project) AddContent(sourceFile string) {
	baseFileName := strings.Split(filepath.Base(sourceFile), ".")[0]
	p.Content = append(p.Content, Content{Name: baseFileName, InFilePath: sourceFile})
}
