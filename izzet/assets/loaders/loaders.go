package loaders

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/kkevinchou/izzet/internal/modelspec"
	"github.com/kkevinchou/izzet/internal/utils"
	"github.com/kkevinchou/izzet/izzet/assets/fonts"
	"github.com/kkevinchou/izzet/izzet/assets/loaders/backends/opengl"
	"github.com/kkevinchou/izzet/izzet/assets/loaders/gltf"
	"github.com/kkevinchou/izzet/izzet/assets/textures"
)

type TextureLoadJob struct {
	key  string
	file string
}

type LoadedTexture struct {
	key  string
	info opengl.TextureInfo
}

func LoadTextures(directory string) map[string]*textures.Texture {
	var subDirectories []string = []string{"images", "icons", "test"}

	extensions := map[string]any{
		".png":  nil,
		".jpeg": nil,
		".jpg":  nil,
	}

	var subPaths []string
	for _, subDir := range subDirectories {
		subPaths = append(subPaths, path.Join(directory, subDir))
	}
	if len(subPaths) == 0 {
		subPaths = append(subPaths, directory)
	}

	fileMetaData := map[string]utils.FileMetaData{}
	for _, subDir := range subPaths {
		utils.GetFileMetaDataRecursive(subDir, extensions, "", false, fileMetaData)
	}

	// fileMetaData := utils.GetFileMetaData(directory, subDirectories, extensions)
	gltfFileMetaData := map[string]utils.FileMetaData{}
	utils.GetFileMetaDataRecursive(filepath.Join(directory, "gltf"), extensions, "", true, gltfFileMetaData)

	filesChan := make(chan TextureLoadJob, len(fileMetaData)+len(gltfFileMetaData))
	loadedTextureChan := make(chan LoadedTexture, len(fileMetaData)+len(gltfFileMetaData))

	workerCount := 10
	doneCount := 0
	var doneCountLock sync.Mutex

	for i := 0; i < workerCount; i++ {
		go func(workerIndex int) {
			for job := range filesChan {
				textureInfo := opengl.ReadTextureInfo(job.file)
				loadedTextureChan <- LoadedTexture{key: job.key, info: textureInfo}
			}

			doneCountLock.Lock()
			doneCount += 1
			if doneCount == workerCount {
				close(loadedTextureChan)
			}
			doneCountLock.Unlock()
		}(i)
	}

	for _, metaData := range fileMetaData {
		filesChan <- TextureLoadJob{key: metaData.Name, file: metaData.Path}
	}
	for _, metaData := range gltfFileMetaData {
		filesChan <- TextureLoadJob{key: metaData.Name, file: metaData.Path}
	}
	close(filesChan)

	textureMap := map[string]*textures.Texture{}
	for loadedTexture := range loadedTextureChan {
		textureID := opengl.CreateOpenGLTexture(loadedTexture.info)
		if _, ok := textureMap[loadedTexture.key]; ok {
			panic(fmt.Sprintf("texture with duplicate name %s found", loadedTexture.key))
		}
		textureMap[loadedTexture.key] = &textures.Texture{ID: textureID}
	}

	return textureMap
}

func LoadTexture(filepath string) *textures.Texture {
	textureInfo := opengl.ReadTextureInfo(filepath)
	textureID := opengl.CreateOpenGLTexture(textureInfo)
	return &textures.Texture{ID: textureID}
}

func LoadDocuments(directory string) map[string]*modelspec.Document {
	var subDirectories []string = []string{"gltf"}

	extensions := map[string]any{
		".gltf": nil,
	}

	scenes := map[string]*modelspec.Document{}
	fileMetaData := utils.GetFileMetaData(directory, subDirectories, extensions)

	for _, metaData := range fileMetaData {
		if strings.HasPrefix(metaData.Name, "_") {
			continue
		}

		if metaData.Extension == ".gltf" {
			scene := LoadDocument(metaData.Name, metaData.Path)
			scenes[metaData.Name] = scene
		} else {
			panic(fmt.Sprintf("wtf unexpected extension %s", metaData.Extension))
		}
	}

	return scenes
}

func LoadDocument(name string, filepath string) *modelspec.Document {
	document, err := gltf.ParseGLTF(name, filepath, &gltf.ParseConfig{TextureCoordStyle: gltf.TextureCoordStyleOpenGL})
	if err != nil {
		panic(err)
	}
	return document
}

func LoadFonts(directory string) map[string]fonts.Font {
	var subDirectories []string = []string{"fonts"}

	extensions := map[string]any{
		".ttf": nil,
	}

	fonts := map[string]fonts.Font{}
	fileMetaData := utils.GetFileMetaData(directory, subDirectories, extensions)

	for _, metaData := range fileMetaData {
		if strings.HasPrefix(metaData.Name, "_") {
			continue
		}
		fonts[metaData.Name] = opengl.NewFont(metaData.Path, 12)
	}

	return fonts
}
