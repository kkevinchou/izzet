package modellibrary

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/model"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/kitolib/modelspec"
	"github.com/kkevinchou/kitolib/utils"
)

type Handle struct {
	namespace string
	id        int
}

func NewHandle(namespace string, id int) *Handle {
	return &Handle{namespace: namespace, id: id}
}

// Interface
// - stores models that have been loaded and builds their VAOs for later rendering
// - models can be referenced via string handle to fetch the associated model / vaos
// - this is used for rendering instantiated entities as well as serialization/deserialization

// Asset Manager
// - represents the model data from disk, agnostic of rendering backend (OpenGL, Vulkan, etc)

// Model Library
// - Multiple implementations, could be backed by OpenGL, Vulkan, etc

// Open Questions
// - is what's the source of truth for mesh data? the asset manager or the model library
// - hierarchical information related to parenting is good hinting that if we construct
// 		the parent entity, we should construct the child entities as well.
//			- e.g. if we construct the base of a house and the roof and walls should come
//				with it as well (since it's parented to the base)
//			- where should this hierarchical information be stored though?

// Notes
// - nodes
//		- have a name
// 		- references a single mesh
// - meshes
//		- have a name
// 		- references a collection of primitives
// - primitives
//		- have a name (not actually useful?)
// 		- VAOs are constructed at the primitive level
//		- each can have their own vertex attributes, materials, etc
// - primitives don't exist in a vacuumm and are always associated with a mesh
// - the model library should be a mesh library using mesh handles (their nane)

type Primitive struct {
	// Name      string
	Primitive *modelspec.PrimitiveSpecification
	Transform mgl32.Mat4

	// vao that contains all vertex attributes
	// position, normals, texture coords, joint indices/weights, etc
	VAO uint32

	// vao that only contains geometry related vertex attributes
	// i.e. vertex positions and joint indices / weights
	// but not normals, texture coords
	GeometryVAO uint32
}

type ModelLibrary struct {
	Primitives map[Handle][]Primitive
	Meshes     map[Handle][]modelspec.MeshSpecification
	Animations map[string]map[string]*modelspec.AnimationSpec
	Joints     map[string]map[int]*modelspec.JointSpec
}

func New() *ModelLibrary {
	m := &ModelLibrary{
		Primitives: map[Handle][]Primitive{},
		Animations: map[string]map[string]*modelspec.AnimationSpec{},
		Joints:     map[string]map[int]*modelspec.JointSpec{},
	}

	m.RegisterMesh("global", cube())

	return m
}

func (m *ModelLibrary) RegisterMesh(namespace string, mesh *modelspec.MeshSpecification) {
	modelConfig := &model.ModelConfig{MaxAnimationJointWeights: settings.MaxAnimationJointWeights}
	vaos := createVAOs(modelConfig, []*modelspec.MeshSpecification{mesh})
	geometryVAOs := createGeometryVAOs(modelConfig, []*modelspec.MeshSpecification{mesh})

	handle := NewHandle(namespace, mesh.ID)
	for i, primitive := range mesh.Primitives {
		m.Primitives[*handle] = append(m.Primitives[*handle], Primitive{
			Primitive:   primitive,
			VAO:         vaos[0][i],
			GeometryVAO: geometryVAOs[0][i],
		})
	}
}

func (m *ModelLibrary) RegisterAnimations(handle string, animations map[string]*modelspec.AnimationSpec, joints map[int]*modelspec.JointSpec) {
	m.Animations[handle] = animations
	m.Joints[handle] = joints
}

func (m *ModelLibrary) GetAnimations(handle string) (map[string]*modelspec.AnimationSpec, map[int]*modelspec.JointSpec) {
	return m.Animations[handle], m.Joints[handle]
}

func (m *ModelLibrary) GetPrimitives(handle Handle) []Primitive {
	return m.Primitives[handle]
}

// maybe this should be computed once and shared across all instances of the mesh?
func UniqueVerticesFromPrimitives(primitives []Primitive) []mgl64.Vec3 {
	var result []mgl64.Vec3
	for _, p := range primitives {
		result = append(result, utils.ModelSpecVertsToVec3(p.Primitive.UniqueVertices)...)
	}
	return result
}
