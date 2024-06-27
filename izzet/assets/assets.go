package assets

import (
	"fmt"
	"time"

	"github.com/kkevinchou/izzet/izzet/assets/assetslog"
	"github.com/kkevinchou/izzet/izzet/assets/fonts"
	"github.com/kkevinchou/izzet/izzet/assets/loaders"
	"github.com/kkevinchou/izzet/izzet/assets/textures"
	"github.com/kkevinchou/kitolib/modelspec"
)

type AssetManager struct {
	textures  map[string]*textures.Texture
	documents map[string]*modelspec.Document
	fonts     map[string]fonts.Font
}

func NewAssetManager(directory string, loadVisualAssets bool) *AssetManager {
	var loadedTextures map[string]*textures.Texture
	var loadedFonts map[string]fonts.Font
	var textureLoadTime time.Duration

	if loadVisualAssets {
		start := time.Now()
		loadedTextures = loaders.LoadTextures(directory)
		textureLoadTime = time.Since(start)
		loadedFonts = loaders.LoadFonts(directory)
	}

	start := time.Now()
	documents := loaders.LoadDocuments(directory)
	assetslog.Logger.Println(textureLoadTime, "to load textures")
	assetslog.Logger.Println(time.Since(start), "to load models")

	assetManager := AssetManager{
		textures:  loadedTextures,
		documents: documents,
		fonts:     loadedFonts,
	}

	return &assetManager
}

func (a *AssetManager) GetTexture(name string) *textures.Texture {
	if _, ok := a.textures[name]; !ok {
		panic(fmt.Sprintf("could not find texture %s", name))
	}
	return a.textures[name]
}

func (a *AssetManager) GetDocument(name string) *modelspec.Document {
	if _, ok := a.documents[name]; !ok {
		panic(fmt.Sprintf("could not find animated model %s", name))
	}
	return a.documents[name]
}

func (a *AssetManager) GetFont(name string) fonts.Font {
	if _, ok := a.fonts[name]; !ok {
		panic(fmt.Sprintf("could not find font %s", name))
	}
	return a.fonts[name]
}

func (a *AssetManager) LoadDocument(name string, filepath string) bool {
	scene := loaders.LoadDocument(name, filepath)
	if _, ok := a.documents[name]; ok {
		fmt.Printf("warning, document with name %s already previously loaded", name)
	}

	a.documents[name] = scene
	return true
}
