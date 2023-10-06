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
	Namespace string
	ID        int
}

func NewHandle(namespace string, id int) Handle {
	return Handle{Namespace: namespace, ID: id}
}

// TODO - need to answer questions around how we know what mesh data to reference when spawning an entity
//		- ideally we have a static and typed handle that we can easily reference from anywhere in the code
//		- this handle should be all we need to construct the mesh component
//		- the mesh component should be all we need to render entities in renderutils
//		- the handle should return all the primitives as well as the animations if any
//		- we need config to be able to mark a document as a single entity that's animated
//		- the registration API for ModelLibrary may need to be a whole document
//		- then the config determines what handle we want to associate with each asset
//			- Question, do I want to support selected instantiation of entities within a document?
//			- e.g. from within demo_scene_samurai, instantiating one entity by name

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
		m.Primitives[handle] = append(m.Primitives[handle], Primitive{
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
