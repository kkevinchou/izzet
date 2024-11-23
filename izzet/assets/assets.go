package assets

import (
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

type AssetManager struct {
	// Static Assets
	textures  map[string]*textures.Texture
	documents map[string]*modelspec.Document
	fonts     map[string]fonts.Font

	// Asset References
	Primitives map[types.MeshHandle][]Primitive
	Animations map[string]map[string]*modelspec.AnimationSpec
	Joints     map[string]map[int]*modelspec.JointSpec
	RootJoints map[string]int
	Materials  map[types.MaterialHandle]modelspec.MaterialSpecification

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
	documents := loaders.LoadDocuments(directory)
	assetslog.Logger.Println(textureLoadTime, "to load textures")
	assetslog.Logger.Println(time.Since(start), "to load models")

	assetManager := AssetManager{
		textures:       loadedTextures,
		documents:      documents,
		fonts:          loadedFonts,
		Primitives:     map[types.MeshHandle][]Primitive{},
		Materials:      map[types.MaterialHandle]modelspec.MaterialSpecification{},
		Animations:     map[string]map[string]*modelspec.AnimationSpec{},
		Joints:         map[string]map[int]*modelspec.JointSpec{},
		RootJoints:     map[string]int{},
		processVisuals: processVisualAssets,
	}

	if processVisualAssets {
		handle := assetManager.GetCubeMeshHandle()
		assetManager.registerMeshWithHandle(handle, cubeMesh(15))

		// default material
		defaultMaterialHandle := assetManager.GetDefaultMaterialHandle()
		assetManager.Materials[defaultMaterialHandle] = modelspec.MaterialSpecification{
			PBRMaterial: &modelspec.PBRMaterial{PBRMetallicRoughness: &modelspec.PBRMetallicRoughness{BaseColorTextureName: settings.DefaultTexture}},
		}

		// TODO: CHEATING - set up all materials prior to registering meshes
		// this will not work for dynamically loaded assets. only the ones loaded on startup

		for _, document := range assetManager.documents {
			for _, material := range document.Materials {
				handle := types.MaterialHandle{ID: material.ID, Namespace: document.Name}
				assetManager.Materials[handle] = material
			}
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
