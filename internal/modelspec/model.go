package modelspec

import (
	"github.com/go-gl/mathgl/mgl32"
)

type PBRMetallicRoughness struct {
	BaseColorTextureIndex       int
	BaseColorTextureName        string
	BaseColorFactor             mgl32.Vec4
	MetalicFactor               float32
	RoughnessFactor             float32
	BaseColorTextureCoordsIndex int
}

type AlphaMode int

const (
	AlphaModeOpaque AlphaMode = 0
	AlphaModeMask   AlphaMode = 1
	AlphaModeBlend  AlphaMode = 2
)

type PBRMaterial struct {
	PBRMetallicRoughness PBRMetallicRoughness
	AlphaMode            AlphaMode
}

type Vertex struct {
	Position       mgl32.Vec3
	Normal         mgl32.Vec3
	Texture0Coords mgl32.Vec2
	Texture1Coords mgl32.Vec2

	JointIDs     []int
	JointWeights []float32
}

type Primitive struct {
	VertexIndices []uint32
	// the unique vertices in the mesh chunk. VertexIndices details
	// how the unique vertices are arranged to construct the mesh
	UniqueVertices []Vertex

	// the ordered vertices where each triplet forms a triangle for the mesh
	Vertices []Vertex

	MaterialIndex *int
}

type Material struct {
	ID          string
	PBRMaterial PBRMaterial
}

type Scene struct {
	Nodes []*Node
}

type Node struct {
	Name      string
	MeshID    *int
	Transform mgl32.Mat4
	Children  []*Node

	Translation mgl32.Vec3
	Rotation    mgl32.Quat
	Scale       mgl32.Vec3
}

type Document struct {
	Name string

	Scenes    []*Scene
	Meshes    []*Mesh
	Materials []Material
	Textures  []string

	JointMap   map[int]*Joint
	Animations map[string]*AnimationSpec

	// not sure where to put this
	RootJoint *Joint

	PeripheralFiles []string
}

type Mesh struct {
	ID         int
	Primitives []*Primitive
}

func NormalizeWeights(jointWeights []JointWeight) {
	var totalWeight float32
	for _, jw := range jointWeights {
		totalWeight += jw.Weight
	}

	for i := range jointWeights {
		jointWeights[i].Weight /= totalWeight
	}
}

type byWeights []JointWeight

type JointWeight struct {
	JointID int
	Weight  float32
}

func (s byWeights) Len() int {
	return len(s)
}
func (s byWeights) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s byWeights) Less(i, j int) bool {
	return s[i].Weight < s[j].Weight
}
