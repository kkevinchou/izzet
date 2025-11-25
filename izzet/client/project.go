package client

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/apputils"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/client/editorcamera"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/world"
)

// Project contains engine data that's meant to be persisted and can be reloaded
type Project struct {
	WorldFile  string
	AssetsFile string
	Name       string
}

type DocumentJSON struct {
	DocumentAsset assets.DocumentAsset
}

type MaterialsJSON struct {
	MaterialAsset assets.MaterialAsset
}

type AssetsJSON struct {
	Documents []DocumentJSON
	Materials []MaterialsJSON
}

func (g *Client) InitializeProjectFolders(name string) error {
	if name == "" {
		return errors.New("name cannot be empty string")
	}

	// project folder
	err := os.MkdirAll(filepath.Join(settings.ProjectsDirectory, name), os.ModePerm)
	if err != nil {
		return err
	}

	// content directory

	err = os.MkdirAll(getContentDir(name), os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func getContentDir(name string) string {
	return filepath.Join(settings.ProjectsDirectory, name, "content")
}

func (g *Client) SaveProject() error {
	return g.SaveProjectAs(g.project.Name)
}

func (g *Client) SaveProjectAs(name string) error {
	if g.project.Name != name {
		err := g.InitializeProjectFolders(name)
		if err != nil {
			return nil
		}
		g.project.Name = name
	}

	contentDir := getContentDir(name)
	worldFilePath := path.Join(settings.ProjectsDirectory, name, "world.json")
	g.saveWorld(worldFilePath)
	g.project.WorldFile = worldFilePath

	assetsJSON := AssetsJSON{}

	// documents
	for _, document := range g.AssetManager().GetDocuments() {
		config := document.Config
		sourceRootDir := filepath.Dir(config.FilePath)

		// don't need to copy assets into the project directory if
		// we loaded it from there

		sourceFilePaths := []string{config.FilePath}
		for _, peripheralFilePath := range document.Document.PeripheralFiles {
			sourceFilePaths = append(sourceFilePaths, filepath.Join(filepath.Dir(config.FilePath), peripheralFilePath))
		}

		newDocument := document
		// in the event where we're saving a new project from the original, we want to overwrite the
		// filepaths so that we reference the documents in the new project directory and not the old one.
		newDocument.Config.FilePath = filepath.Join(contentDir, filepath.Base(config.FilePath))
		assetsJSON.Documents = append(assetsJSON.Documents, DocumentJSON{
			DocumentAsset: newDocument,
		})

		if sourceRootDir == contentDir {
			continue
		}

		err := copySourceFiles(sourceFilePaths, sourceRootDir, contentDir)
		if err != nil {
			panic(err)
		}
	}

	// materials

	for _, material := range g.AssetManager().GetMaterials() {
		assetsJSON.Materials = append(assetsJSON.Materials, MaterialsJSON{MaterialAsset: material})
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

	g.project.AssetsFile = assetsFilePath

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

func (g *Client) NewProject(name string) {
	g.InitializeProjectFolders(name)
	g.project = &Project{Name: name}
	g.assetManager = assets.NewAssetManager(true)
	g.world = world.New()

	g.initializeAppSystems()
	g.LoadDefaultAssets()
	g.SelectEntity(nil)

	// set up the default scene

	g.camera = &editorcamera.Camera{
		Position: settings.EditorCameraStartPosition,
		Rotation: mgl64.QuatIdent(),
	}

	cube := entity.CreateCube(g.AssetManager(), 1)
	cube.Material = &entity.MaterialComponent{MaterialHandle: assets.DefaultMaterialHandle}
	entity.SetLocalPosition(cube, mgl64.Vec3{0, -1, 0})
	entity.SetScale(cube, mgl64.Vec3{7, 2, 7})
	g.World().AddEntity(cube)

	directionalLight := entity.CreateDirectionalLight()
	directionalLight.LightInfo.Diffuse3F = [3]float32{1, 1, 1}
	directionalLight.LightInfo.Direction3F = [3]float32{-0.5, -1, 1}
	directionalLight.Name = "directional_light"
	directionalLight.LightInfo.PreScaledIntensity = 4
	entity.SetLocalPosition(directionalLight, mgl64.Vec3{0, 20, 0})
	g.World().AddEntity(directionalLight)

	g.SaveProjectAs(name)
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

	g.project = &project
	g.world = world.New()
	g.initializeAppSystems()
	g.RuntimeConfig().BatchRenderingEnabled = false

	g.initializeAssetManagerWithProject(name)
	success := g.initializeAppAndWorldFromFile(path.Join(settings.ProjectsDirectory, name, "world.json"))

	return success
}

func (g *Client) initializeAssetManagerWithProject(name string) {
	assetsFilePath := path.Join(settings.ProjectsDirectory, name, "assets.json")
	_, err := os.Stat(assetsFilePath)
	if err != nil {
		panic(err)
	}

	assetsFile, err := os.Open(assetsFilePath)
	if err != nil {
		panic(err)
	}
	defer assetsFile.Close()

	var assetsJSON AssetsJSON
	decoder := json.NewDecoder(assetsFile)
	err = decoder.Decode(&assetsJSON)
	if err != nil {
		panic(err)
	}

	g.assetManager = assets.NewAssetManager(true)

	// load meshes, skip materials
	// materials are skipped because the materials from the document should already
	// be saved to our assets.json file which is loaded independently

	// TODO - take assetJSON as the input?
	for _, document := range assetsJSON.Documents {
		// TODO - issue: the document in document asset is populated by reading the config
		// which still points to the default location for assets. actually is this an issue?
		// ideally all assets are pointed to the .project folder. perhaps we should have a new project
		// setup step that copies assets over from _assets rather than loading directly from _assets?
		g.assetManager.LoadAndRegisterDocumentAsset(document.DocumentAsset)
	}

	for _, material := range assetsJSON.Materials {
		g.assetManager.CreateMaterialWithHandle(material.MaterialAsset.Name, material.MaterialAsset.Material, material.MaterialAsset.Handle)
	}
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
