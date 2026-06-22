package assets

import (
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/iztlog"
	"github.com/kkevinchou/izzet/internal/modelspec"
	"github.com/kkevinchou/izzet/internal/platforms"
	"github.com/kkevinchou/izzet/internal/utils"
	"github.com/kkevinchou/izzet/izzet/assets/fonts"
	"github.com/kkevinchou/izzet/izzet/assets/loaders"
	"github.com/kkevinchou/izzet/izzet/assets/textures"
	"github.com/kkevinchou/izzet/izzet/settings"
)

var materialIDGen int = 0
var izzetMaterialPrefix = "izzet/"
var fallbackTexture string = "default"

type Document struct {
	MatIDToHandle map[string]MaterialHandle
	Document      *modelspec.Document `json:"-"`
	Config        AssetConfig
}

type Material struct {
	Material modelspec.Material
	Name     string
	Handle   MaterialHandle
}

type Primitive struct {
	Primitive *modelspec.Primitive

	// vao that contains all vertex attributes
	// position, normals, texture coords, joint indices/weights, etc
	VAO uint32

	// vao that only contains geometry related vertex attributes
	// i.e. vertex positions and joint indices / weights
	// but not normals, texture coords
	GeometryVAO uint32

	MaterialHandle MaterialHandle
}

type AssetManager struct {
	logger *slog.Logger
	// Static Assets
	textures  map[string]*textures.Texture
	documents map[string]Document
	fonts     map[string]fonts.Font
	materials map[MaterialHandle]Material

	// Asset References
	Primitives map[MeshHandle][]Primitive
	Animations map[string]map[string]*modelspec.AnimationSpec
	Joints     map[string]map[int]*modelspec.Joint
	RootJoints map[string]int

	processVisuals bool
	audioData      map[string]loaders.AudioData
}

func NewAssetManager(processVisualAssets bool, logger *slog.Logger) *AssetManager {
	var loadedTextures map[string]*textures.Texture
	var loadedFonts map[string]fonts.Font

	var audioData map[string]loaders.AudioData

	if processVisualAssets {
		start := time.Now()
		loadedTextures = loaders.LoadTextures(settings.BuiltinAssetsDir)
		loadedFonts = loaders.LoadFonts(settings.BuiltinAssetsDir)

		audioData = loaders.LoadAudio(settings.BuiltinAssetsDir, platforms.AudioMixer())
		iztlog.ClientLogger.Info(fmt.Sprintf("loaded fonts and textures in %f seconds", time.Since(start).Seconds()))
	}

	assetManager := AssetManager{
		logger:         logger,
		textures:       loadedTextures,
		fonts:          loadedFonts,
		audioData:      audioData,
		processVisuals: processVisualAssets,
		documents:      map[string]Document{},
		Primitives:     map[MeshHandle][]Primitive{},
		materials:      map[MaterialHandle]Material{},
		Animations:     map[string]map[string]*modelspec.AnimationSpec{},
		Joints:         map[string]map[int]*modelspec.Joint{},
		RootJoints:     map[string]int{},
	}

	assetManager.registerMeshPrimitivesWithHandle(defaultCubeHandle, CreateCubeMesh(1), nil)

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

func (a *AssetManager) Play(name string) {
	if audioData, ok := a.audioData[name]; ok {
		if err := audioData.Play(); err != nil {
			panic(fmt.Errorf("play audio %s: %w", name, err))
		}
	}
}

func (a *AssetManager) GetTexture(name string) *textures.Texture {
	if _, ok := a.textures[name]; !ok {
		panic(fmt.Sprintf("could not find texture %s", name))
	}
	return a.textures[name]
}

func (a *AssetManager) GetTextureWithFallback(name string) *textures.Texture {
	if _, ok := a.textures[name]; !ok {
		return a.GetTexture(fallbackTexture)
	}
	return a.textures[name]
}

func (a *AssetManager) GetDocumentAsset(name string) Document {
	if _, ok := a.documents[name]; !ok {
		panic(fmt.Sprintf("could not find animated model %s", name))
	}
	return a.documents[name]
}

func (a *AssetManager) GetDocument(name string) *modelspec.Document {
	if _, ok := a.documents[name]; !ok {
		panic(fmt.Sprintf("could not find animated model %s", name))
	}
	return a.documents[name].Document
}

func (a *AssetManager) GetDocuments() []Document {
	var documents []Document
	for _, d := range a.documents {
		documents = append(documents, d)
	}
	sort.Slice(documents, func(i, j int) bool {
		return documents[i].Config.Name < documents[j].Config.Name
	})
	return documents
}

func (a *AssetManager) GetMaterials() []Material {
	var materials []Material
	for _, material := range a.materials {
		materials = append(materials, material)
	}
	sort.Slice(materials, func(i, j int) bool {
		return materials[i].Name < materials[j].Name
	})
	return materials
}

func (m *AssetManager) GetMaterial(materialHandle MaterialHandle) Material {
	if materialAsset, ok := m.materials[materialHandle]; ok {
		return materialAsset
	}
	material := m.materials[defaultMaterialHandle]
	return material
}

func (m *AssetManager) DeleteMaterial(materialHandle MaterialHandle) {
	delete(m.materials, materialHandle)
}

func (m *AssetManager) UpdateMaterialAsset(material Material) {
	if _, ok := m.materials[material.Handle]; ok {
		m.materials[material.Handle] = material
		return
	}
	panic("material handle not found")
}

func (m *AssetManager) CreateCustomMaterial(name string, material modelspec.Material) MaterialHandle {
	materialHandle := MaterialHandle{id: fmt.Sprintf("%s%d", izzetMaterialPrefix, materialIDGen)}
	if mat, ok := m.materials[materialHandle]; ok {
		panic(fmt.Sprintf("material already exists in asset manager. %v", mat))
	}
	materialIDGen++
	m.materials[materialHandle] = Material{Material: material, Handle: materialHandle, Name: name}
	return materialHandle
}

func (m *AssetManager) createMaterial(name string, id string, material modelspec.Material) MaterialHandle {
	materialHandle := MaterialHandle{id: id}
	m.materials[materialHandle] = Material{Material: material, Handle: materialHandle, Name: name}
	return materialHandle
}

func (m *AssetManager) CreateMaterialWithHandle(name string, material modelspec.Material, materialHandle MaterialHandle) {
	if _, ok := m.materials[materialHandle]; !ok {
		m.materials[materialHandle] = Material{Material: material, Handle: materialHandle, Name: name}
	}
	materialID := string(materialHandle.id)
	if strings.HasPrefix(materialID, izzetMaterialPrefix) {
		// this is an ugly hack, pls fix
		split := strings.Split(materialID, "/")
		if len(split) == 2 {
			if id, err := strconv.Atoi(split[1]); err == nil {
				if materialIDGen <= id {
					materialIDGen = id + 1
				}
			}
		}
	}
}

func (a *AssetManager) GetFont(name string) fonts.Font {
	if _, ok := a.fonts[name]; !ok {
		panic(fmt.Sprintf("could not find font %s", name))
	}
	return a.fonts[name]
}

func (a *AssetManager) DeleteDocument(document Document) {
	delete(a.documents, document.Config.Name)
}
