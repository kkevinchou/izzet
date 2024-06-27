package loaders

import (
	"fmt"
	"strings"
	"sync"

	"github.com/kkevinchou/izzet/izzet/assets/fonts"
	"github.com/kkevinchou/izzet/izzet/assets/loaders/backends/opengl"
	"github.com/kkevinchou/izzet/izzet/assets/loaders/gltf"
	"github.com/kkevinchou/izzet/izzet/assets/textures"
	"github.com/kkevinchou/kitolib/modelspec"
	"github.com/kkevinchou/kitolib/utils"
)

func LoadTextures(directory string) map[string]*textures.Texture {
	var subDirectories []string = []string{"images", "icons", "gltf", "test"}

	extensions := map[string]any{
		".png":  nil,
		".jpeg": nil,
		".jpg":  nil,
	}

	fileMetaData := utils.GetFileMetaData(directory, subDirectories, extensions)

	filesChan := make(chan string, len(fileMetaData))
	textureInfoChan := make(chan opengl.TextureInfo, len(fileMetaData))

	workerCount := 10
	doneCount := 0
	var doneCountLock sync.Mutex

	for i := 0; i < workerCount; i++ {
		go func(workerIndex int) {
			for fileName := range filesChan {
				textureInfo := opengl.ReadTextureInfo(fileName)
				textureInfoChan <- textureInfo
			}

			doneCountLock.Lock()
			doneCount += 1
			if doneCount == workerCount {
				close(textureInfoChan)
			}
			doneCountLock.Unlock()
		}(i)
	}

	for _, metaData := range fileMetaData {
		filesChan <- metaData.Path
	}
	close(filesChan)

	textureMap := map[string]*textures.Texture{}
	for textureInfo := range textureInfoChan {
		textureID := opengl.CreateOpenGLTexture(textureInfo)
		if _, ok := textureMap[textureInfo.Name]; ok {
			panic(fmt.Sprintf("texture with duplicate name %s found", textureInfo.Name))
		}
		textureMap[textureInfo.Name] = &textures.Texture{ID: textureID}
	}

	return textureMap
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
