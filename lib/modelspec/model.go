package modelspec

import "github.com/go-gl/mathgl/mgl32"

type PBRMetallicRoughness struct {
	BaseColorTextureIndex *int
	BaseColorTextureName  string
	BaseColorFactor       mgl32.Vec4
	MetalicFactor         float32
	RoughnessFactor       float32
}

type PBRMaterial struct {
	PBRMetallicRoughness *PBRMetallicRoughness
}

type Vertex struct {
	Position mgl32.Vec3
	Normal   mgl32.Vec3
	Texture  mgl32.Vec2

	JointIDs     []int
	JointWeights []float32
}

type MeshChunkSpecification struct {
	VertexIndices []uint32
	// the unique vertices in the mesh chunk. VertexIndices details
	// how the unique vertices are arranged to construct the mesh
	UniqueVertices []Vertex

	// the ordered vertices where each triplet forms a triangle for the mesh
	Vertices []Vertex

	// PBR
	PBRMaterial *PBRMaterial
}
type MeshSpecification struct {
	MeshChunks []*MeshChunkSpecification
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

	// list of textures by name. the index within this slice is
	// the id for which the modespec references textures
	Textures []string
}
