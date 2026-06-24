package assets

import (
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	"github.com/kkevinchou/izzet/internal/iztlog"
	"github.com/kkevinchou/izzet/internal/modelspec"
	"github.com/kkevinchou/izzet/internal/platforms"
	"github.com/kkevinchou/izzet/internal/utils"
	"github.com/kkevinchou/izzet/izzet/assets/fonts"
	"github.com/kkevinchou/izzet/izzet/assets/loaders"
	"github.com/kkevinchou/izzet/izzet/assets/textures"
	"github.com/kkevinchou/izzet/izzet/settings"
)

var fallbackTexture string = "default"

type Document struct {
	// mapping from the source document material ID to the in-engine material ids
	SourceMaterialIDToMaterialID map[string]MaterialID
	Document                     *modelspec.Document `json:"-"`
	ID                           string
	Filepath                     string
}

type Material struct {
	Material modelspec.Material
	Name     string
	ID       MaterialID
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

	MaterialID MaterialID
}

type AssetManager struct {
	logger *slog.Logger
	// Static Assets
	textures  map[string]*textures.Texture
	documents map[string]Document
	fonts     map[string]fonts.Font
	materials map[MaterialID]Material

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
		materials:      map[MaterialID]Material{},
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
		return documents[i].ID < documents[j].ID
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

func (m *AssetManager) GetMaterial(materialID MaterialID) Material {
	if materialAsset, ok := m.materials[materialID]; ok {
		return materialAsset
	}
	material := m.materials[defaultMaterialID]
	return material
}

func (m *AssetManager) DeleteMaterial(materialID MaterialID) {
	delete(m.materials, materialID)
}

func (m *AssetManager) UpdateMaterialAsset(material Material) {
	if _, ok := m.materials[material.ID]; ok {
		m.materials[material.ID] = material
		return
	}
	panic("material id not found")
}

func (m *AssetManager) CreateCustomMaterial(name string, material modelspec.Material) MaterialID {
	materialID := MaterialID("material:" + uuid.NewString())
	if mat, ok := m.materials[materialID]; ok {
		panic(fmt.Sprintf("material already exists in asset manager. %v", mat))
	}
	m.materials[materialID] = Material{Material: material, ID: materialID, Name: name}
	return materialID
}

func (m *AssetManager) createMaterial(name string, id MaterialID, material modelspec.Material) MaterialID {
	m.materials[id] = Material{Material: material, ID: id, Name: name}
	return id
}

func (m *AssetManager) CreateMaterialWithID(name string, material modelspec.Material, materialID MaterialID) {
	if _, ok := m.materials[materialID]; !ok {
		m.materials[materialID] = Material{Material: material, ID: materialID, Name: name}
	}
}

func (a *AssetManager) GetFont(name string) fonts.Font {
	if _, ok := a.fonts[name]; !ok {
		panic(fmt.Sprintf("could not find font %s", name))
	}
	return a.fonts[name]
}

func (a *AssetManager) DeleteDocument(document Document) {
	delete(a.documents, document.ID)
}
