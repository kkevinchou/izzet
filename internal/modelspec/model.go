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

type PrimitiveSpecification struct {
	VertexIndices []uint32
	// the unique vertices in the mesh chunk. VertexIndices details
	// how the unique vertices are arranged to construct the mesh
	UniqueVertices []Vertex

	// the ordered vertices where each triplet forms a triangle for the mesh
	Vertices []Vertex

	MaterialIndex string
}

type MaterialSpecification struct {
	ID          string
	PBRMaterial PBRMaterial
}

// ModelSpecification is the output of any parsed model files (e.g. from Blender, Maya, etc)
// and acts a the blueprint for the model that contains all the associated vertex and
// animation data. This struct should be agnostic to the 3D modelling tool that produced the data.
type ModelSpecification struct {
	Meshes []*MeshSpecification

	// Joint Hierarchy
	RootJoint *JointSpec

	// Animations
	Animations map[string]*AnimationSpec

	RootTransforms mgl32.Mat4

	JointMap map[int]*JointSpec

	// list of textures by name. the index within this slice is
	// the id for which the modespec references textures
	Textures []string
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
	Meshes    []*MeshSpecification
	Materials []MaterialSpecification
	Textures  []string

	JointMap   map[int]*JointSpec
	Animations map[string]*AnimationSpec

	// not sure where to put this
	RootJoint *JointSpec

	PeripheralFiles []string
}

type MeshSpecification struct {
	ID         int
	Primitives []*PrimitiveSpecification
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

// careful with this method. i believe this assumes that the local bind pose is in a tpose but this isn't always the case.
// in collada files it's more reliable to read the inv bind matrix from the data file itself rather than try to calculate it
// func calculateInverseBindTransform(joint *modelspec.JointSpec, parentBindTransform mgl32.Mat4) {
// 	bindTransform := parentBindTransform.Mul4(joint.BindTransform) // model-space relative to the origin
// 	joint.InverseBindTransform = bindTransform.Inv()
// 	for _, child := range joint.Children {
// 		calculateInverseBindTransform(child, bindTransform)
// 	}
// }
