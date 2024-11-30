package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/kkevinchou/izzet/izzet/apputils"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/settings"
)

// Project contains engine data that's meant to be persisted and can be reloaded
type Project struct {
	WorldFile  string
	AssetsFile string
	Name       string
}

func NewProject() *Project {
	return &Project{}
}

type DocumentJSON struct {
	Config assets.AssetConfig
}

type AssetsJSON struct {
	Documents []DocumentJSON
}

func (g *Client) SaveProject(name string) error {
	if name == "" {
		return errors.New("name cannot be empty string")
	}

	// project folder
	err := os.MkdirAll(filepath.Join(settings.ProjectsDirectory, name), os.ModePerm)
	if err != nil {
		panic(err)
	}

	worldFilePath := path.Join(settings.ProjectsDirectory, name, fmt.Sprintf("./%s.json", name))
	g.saveWorld(worldFilePath)
	g.project.WorldFile = worldFilePath

	// content directory

	contentDir := filepath.Join(settings.ProjectsDirectory, name, "content")
	err = os.MkdirAll(contentDir, os.ModePerm)
	if err != nil {
		panic(err)
	}

	// assets

	assetsJSON := AssetsJSON{}

	for _, document := range g.AssetManager().GetDocuments() {
		config := document.Config
		sourceRootDir := filepath.Dir(config.FilePath)

		// don't need to copy assets into the project directory if
		// we loaded it from there

		sourceFilePaths := []string{config.FilePath}
		for _, peripheralFilePath := range document.Document.PeripheralFiles {
			sourceFilePaths = append(sourceFilePaths, filepath.Join(filepath.Dir(config.FilePath), peripheralFilePath))
		}

		newConfig := document.Config
		newConfig.FilePath = filepath.Join(contentDir, filepath.Base(config.FilePath))
		assetsJSON.Documents = append(assetsJSON.Documents, DocumentJSON{
			Config: newConfig,
		})

		if sourceRootDir == contentDir {
			continue
		}

		err := copySourceFiles(sourceFilePaths, sourceRootDir, contentDir)
		if err != nil {
			panic(err)
		}
	}

	assetsFilePath := path.Join(settings.ProjectsDirectory, name, "assets.json")
	assetsFile, err := os.OpenFile(assetsFilePath, os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer assetsFile.Close()

	encoder := json.NewEncoder(assetsFile)
	encoder.SetIndent("", "    ")
	err = encoder.Encode(assetsJSON)
	if err != nil {
		panic(err)
	}

	f, err := os.OpenFile(filepath.Join(settings.ProjectsDirectory, name, "main_project.izt"), os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	encoder = json.NewEncoder(f)
	encoder.SetIndent("", "    ")
	err = encoder.Encode(g.project)
	if err != nil {
		panic(err)
	}

	return nil
}

func (g *Client) LoadProject(name string) bool {
	if name == "" {
		return false
	}

	f, err := os.Open(apputils.PathToProjectFile(name))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var project Project
	decoder := json.NewDecoder(f)
	err = decoder.Decode(&project)
	if err != nil {
		panic(err)
	}

	// assets

	g.assetManager.Reset()
	assetsFilePath := path.Join(settings.ProjectsDirectory, name, "assets.json")
	_, err = os.Stat(assetsFilePath)
	if err == nil {
		assetsFile, err := os.Open(assetsFilePath)
		if err != nil {
			panic(err)
		}
		defer assetsFile.Close()

		var assetsJSON AssetsJSON
		decoder = json.NewDecoder(assetsFile)
		err = decoder.Decode(&assetsJSON)
		if err != nil {
			panic(err)
		}

		for _, document := range assetsJSON.Documents {
			g.assetManager.LoadAndRegisterDocument(document.Config)
		}
	}

	g.project = &project

	return g.loadWorld(path.Join(settings.ProjectsDirectory, name, name+".json"))
}

func copySourceFiles(filePaths []string, rootDir, dstDir string) error {
	for _, src := range filePaths {
		// Get the relative path of the source file
		relPath, err := filepath.Rel(rootDir, src)
		if err != nil {
			return err
		}

		// Construct the destination path
		destPath := filepath.Join(dstDir, relPath)

		// Check if the source is a directory
		info, err := os.Stat(src)
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Create the directory at the destination
			err = os.MkdirAll(destPath, info.Mode())
			if err != nil {
				return err
			}
		} else {
			// Ensure the destination directory exists
			err = os.MkdirAll(filepath.Dir(destPath), os.ModePerm)
			if err != nil {
				return err
			}

			// Copy the file
			err = copyFile(src, destPath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	// Open the source file
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Create the destination file
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy the contents from source to destination
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Flush the file to ensure all data is written
	err = destFile.Sync()
	if err != nil {
		return err
	}

	return nil
}
