package assets

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/go-gl/mathgl/mgl32"
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

const (
	defaultMaterialName string = "default material"
	whiteMaterialName   string = "white material"
)

type DocumentAsset struct {
	Document *modelspec.Document `json:"-"`
	Config   AssetConfig
}

type MaterialAsset struct {
	Material modelspec.MaterialSpecification
	Name     string
	Handle   types.MaterialHandle
}

type AssetManager struct {
	// Static Assets
	textures       map[string]*textures.Texture
	documentAssets map[string]DocumentAsset
	fonts          map[string]fonts.Font
	materialAssets map[types.MaterialHandle]MaterialAsset

	// Asset References
	NamespaceToMeshHandles map[string][]types.MeshHandle
	Primitives             map[types.MeshHandle][]Primitive
	Animations             map[string]map[string]*modelspec.AnimationSpec
	Joints                 map[string]map[int]*modelspec.JointSpec
	RootJoints             map[string]int

	processVisuals bool
}

func NewAssetManager(processVisualAssets bool) *AssetManager {
	var loadedTextures map[string]*textures.Texture
	var loadedFonts map[string]fonts.Font

	if processVisualAssets {
		start := time.Now()
		loadedTextures = loaders.LoadTextures(settings.BuiltinAssetsDir)
		loadedFonts = loaders.LoadFonts(settings.BuiltinAssetsDir)
		assetslog.Logger.Println("loaded fonts and textures in", time.Since(start).Seconds(), "seconds")
	}

	assetManager := AssetManager{
		textures:       loadedTextures,
		fonts:          loadedFonts,
		processVisuals: processVisualAssets,
	}
	assetManager.Reset()

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

func (a *AssetManager) GetDocumentAsset(name string) DocumentAsset {
	if _, ok := a.documentAssets[name]; !ok {
		panic(fmt.Sprintf("could not find animated model %s", name))
	}
	return a.documentAssets[name]
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
		return documents[i].Config.Name < documents[j].Config.Name
	})
	return documents
}

func (a *AssetManager) GetMaterials() []MaterialAsset {
	var materials []MaterialAsset
	for _, material := range a.materialAssets {
		materials = append(materials, material)
	}
	sort.Slice(materials, func(i, j int) bool {
		return materials[i].Name < materials[j].Name
	})
	return materials
}

func (m *AssetManager) GetMaterial(handle types.MaterialHandle) MaterialAsset {
	if materialAsset, ok := m.materialAssets[handle]; ok {
		return materialAsset
	}
	material := m.materialAssets[DefaultMaterialHandle]
	return material
}

func (m *AssetManager) UpdateMaterialAsset(material MaterialAsset) {
	if _, ok := m.materialAssets[material.Handle]; ok {
		m.materialAssets[material.Handle] = material
		return
	}
	panic(fmt.Sprintf("%s handle not found", material.Handle.String()))
}

var materialIDGen int = 100

func (m *AssetManager) CreateMaterial(name string, material modelspec.MaterialSpecification) types.MaterialHandle {
	handle := NewMaterialHandle(NamespaceGlobal, fmt.Sprintf("%d", materialIDGen))
	materialIDGen++
	return m.CreateMaterialWithHandle(name, material, handle)
}

func (m *AssetManager) CreateMaterialWithHandle(name string, material modelspec.MaterialSpecification, handle types.MaterialHandle) types.MaterialHandle {
	m.materialAssets[handle] = MaterialAsset{Material: material, Handle: handle, Name: name}
	return handle
}

func (m *AssetManager) CreateMaterialWithHandleNoOverride(name string, material modelspec.MaterialSpecification, handle types.MaterialHandle) types.MaterialHandle {
	if _, ok := m.materialAssets[handle]; !ok {
		m.materialAssets[handle] = MaterialAsset{Material: material, Handle: handle, Name: name}
	}
	return handle
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
	a.materialAssets = map[types.MaterialHandle]MaterialAsset{}
	a.Animations = map[string]map[string]*modelspec.AnimationSpec{}
	a.Joints = map[string]map[int]*modelspec.JointSpec{}
	a.RootJoints = map[string]int{}

	if a.processVisuals {
		handle := a.GetCubeMeshHandle()
		a.registerMeshPrimitivesWithHandle(handle, cubeMesh(15))

		// default materials
		material := modelspec.MaterialSpecification{
			PBRMaterial: modelspec.PBRMaterial{
				PBRMetallicRoughness: modelspec.PBRMetallicRoughness{
					BaseColorTextureName: settings.DefaultTexture,
					// BaseColorTextureName: "",
					BaseColorFactor: mgl32.Vec4{1, 1, 1, 1},
					RoughnessFactor: .55,
					MetalicFactor:   0,
				},
			},
		}
		a.CreateMaterialWithHandleNoOverride(defaultMaterialName, material, DefaultMaterialHandle)

		whiteMaterial := modelspec.MaterialSpecification{
			PBRMaterial: modelspec.PBRMaterial{
				PBRMetallicRoughness: modelspec.PBRMetallicRoughness{
					// BaseColorTextureName: settings.DefaultTexture,
					// BaseColorTextureName: "",
					BaseColorFactor: mgl32.Vec4{1, 1, 1, 1},
					RoughnessFactor: .55,
					MetalicFactor:   0,
				},
			},
		}
		a.CreateMaterialWithHandleNoOverride(whiteMaterialName, whiteMaterial, DefaultMaterialHandle)

		// load builtin assets

		var subDirectories []string = []string{"gltf"}
		extensions := map[string]any{
			".gltf": nil,
		}
		fileMetaData := utils.GetFileMetaData(settings.BuiltinAssetsDir, subDirectories, extensions)
		for _, metaData := range fileMetaData {
			if strings.HasPrefix(metaData.Name, "_") {
				continue
			}

			a.LoadAndRegisterDocument(AssetConfig{
				Name:          metaData.Name,
				FilePath:      metaData.Path,
				ColliderType:  string(types.ColliderTypeMesh),
				ColliderGroup: string(types.ColliderGroupPlayer),
				SingleEntity:  true,
			}, true)
		}
	}
}
