package project

import (
	"encoding/json"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/kkevinchou/izzet/izzet/settings"
)

type Content struct {
	Name     string
	Filepath string
}

// type ContentBrowser struct {
// 	Content map[string]Content
// }

type Project struct {
	Content map[string]Content
	// ContentBrowser ContentBrowser
	// World          world.GameWorld
}

func NewProject() *Project {
	return &Project{Content: map[string]Content{}}
}

func (p *Project) AddContent(sourceFile string) {
	baseFileName := strings.Split(filepath.Base(sourceFile), ".")[0]

	importedFile, err := os.Open(sourceFile)
	if err != nil {
		panic(err)
	}
	defer importedFile.Close()

	fileBytes, err := io.ReadAll(importedFile)
	if err != nil {
		panic(err)
	}

	outFile, err := os.OpenFile(path.Join(settings.ProjectDirectory, "content", baseFileName+filepath.Ext(sourceFile)), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	_, err = outFile.Write(fileBytes)
	if err != nil {
		panic(err)
	}

	p.Content[baseFileName] = Content{Name: baseFileName, Filepath: outFile.Name()}
}

func (p *Project) Save() {
	f, err := os.OpenFile(filepath.Join(settings.ProjectDirectory, "main_project.izt"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.Encode(p)
}

// 1. Save the project file
// 		The current level
// 		The content browser
// 2. Load the project file
