package assets

import (
	"fmt"
	"sort"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/assets/assetslog"
	"github.com/kkevinchou/izzet/izzet/assets/fonts"
	"github.com/kkevinchou/izzet/izzet/assets/loaders"
	"github.com/kkevinchou/izzet/izzet/assets/textures"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/kitolib/modelspec"
	"github.com/kkevinchou/kitolib/utils"
)

type DocumentAsset struct {
	Name           string
	Document       *modelspec.Document `json:"-"`
	SourceFilePath string
	Config         AssetConfig
}

type AssetManager struct {
	// Static Assets
	textures       map[string]*textures.Texture
	documentAssets map[string]DocumentAsset
	fonts          map[string]fonts.Font

	// Asset References
	NamespaceToMeshHandles map[string][]types.MeshHandle
	Primitives             map[types.MeshHandle][]Primitive
	Animations             map[string]map[string]*modelspec.AnimationSpec
	Joints                 map[string]map[int]*modelspec.JointSpec
	RootJoints             map[string]int
	Materials              map[types.MaterialHandle]modelspec.MaterialSpecification

	processVisuals bool
}

func NewAssetManager(directory string, processVisualAssets bool) *AssetManager {
	var loadedTextures map[string]*textures.Texture
	var loadedFonts map[string]fonts.Font
	var textureLoadTime time.Duration

	if processVisualAssets {
		start := time.Now()
		loadedTextures = loaders.LoadTextures(directory)
		textureLoadTime = time.Since(start)
		loadedFonts = loaders.LoadFonts(directory)
	}

	start := time.Now()
	// documents := loaders.LoadDocuments(directory)
	assetslog.Logger.Println(textureLoadTime, "to load textures")
	assetslog.Logger.Println(time.Since(start), "to load models")

	assetManager := AssetManager{
		textures:       loadedTextures,
		fonts:          loadedFonts,
		processVisuals: processVisualAssets,
	}
	assetManager.Reset()

	if processVisualAssets {
		handle := assetManager.GetCubeMeshHandle()
		assetManager.registerMeshPrimitivesWithHandle(handle, cubeMesh(15))

		// default material
		defaultMaterialHandle := assetManager.GetDefaultMaterialHandle()
		assetManager.Materials[defaultMaterialHandle] = modelspec.MaterialSpecification{
			PBRMaterial: &modelspec.PBRMaterial{PBRMetallicRoughness: &modelspec.PBRMetallicRoughness{BaseColorTextureName: settings.DefaultTexture}},
		}
	}

	return &assetManager
}

// maybe this should be computed once and shared across all instances of the mesh?
func UniqueVerticesFromPrimitives(primitives []Primitive) []mgl64.Vec3 {
	var result []mgl64.Vec3
	for _, p := range primitives {
		result = append(result, utils.ModelSpecVertsToVec3(p.Primitive.UniqueVertices)...)
	}
	return result
}

func (a *AssetManager) GetTexture(name string) *textures.Texture {
	if _, ok := a.textures[name]; !ok {
		panic(fmt.Sprintf("could not find texture %s", name))
	}
	return a.textures[name]
}

func (a *AssetManager) GetDocument(name string) *modelspec.Document {
	if _, ok := a.documentAssets[name]; !ok {
		panic(fmt.Sprintf("could not find animated model %s", name))
	}
	return a.documentAssets[name].Document
}

func (a *AssetManager) GetDocuments() []DocumentAsset {
	var documents []DocumentAsset
	for _, documentAsset := range a.documentAssets {
		documents = append(documents, documentAsset)
	}
	sort.Slice(documents, func(i, j int) bool {
		return documents[i].Name < documents[j].Name
	})
	return documents
}

func (a *AssetManager) GetFont(name string) fonts.Font {
	if _, ok := a.fonts[name]; !ok {
		panic(fmt.Sprintf("could not find font %s", name))
	}
	return a.fonts[name]
}

func (a *AssetManager) Reset() {
	a.documentAssets = map[string]DocumentAsset{}
	a.Primitives = map[types.MeshHandle][]Primitive{}
	a.NamespaceToMeshHandles = map[string][]types.MeshHandle{}
	a.Materials = map[types.MaterialHandle]modelspec.MaterialSpecification{}
	a.Animations = map[string]map[string]*modelspec.AnimationSpec{}
	a.Joints = map[string]map[int]*modelspec.JointSpec{}
	a.RootJoints = map[string]int{}
}
