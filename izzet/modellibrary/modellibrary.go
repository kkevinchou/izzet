package modellibrary

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/kitolib/modelspec"
	"github.com/kkevinchou/kitolib/utils"
)

var nextGlobalID int

const (
	NamespaceGlobal = "global"

	HandleTypeCube HandleType = "cube"
)

type HandleType string

type Handle struct {
	Namespace string
	ID        string

	Type         HandleType
	IntTypeParam int
}

type ModelConfig struct {
	MaxAnimationJointWeights int
}

func NewGlobalHandle(id string) Handle {
	return Handle{Namespace: NamespaceGlobal, ID: id}
}

func NewHandle(namespace string, id string) Handle {
	return Handle{Namespace: namespace, ID: id}
}

func NewHandleFromMeshID(namespace string, meshID int) Handle {
	return NewHandle(namespace, fmt.Sprintf("%d", meshID))
}

type Primitive struct {
	Primitive *modelspec.PrimitiveSpecification

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
	Animations map[string]map[string]*modelspec.AnimationSpec
	Joints     map[string]map[int]*modelspec.JointSpec
	RootJoints map[string]int

	processVisuals bool
}

func New(processVisuals bool) *ModelLibrary {
	m := &ModelLibrary{
		Primitives:     map[Handle][]Primitive{},
		Animations:     map[string]map[string]*modelspec.AnimationSpec{},
		Joints:         map[string]map[int]*modelspec.JointSpec{},
		RootJoints:     map[string]int{},
		processVisuals: processVisuals,
	}

	return m
}

func (m *ModelLibrary) GetOrCreateCubeMeshHandle(length int) Handle {
	handle := NewHandle("global", fmt.Sprintf("cube-%d", length))
	handle.Type = HandleTypeCube
	handle.IntTypeParam = length
	if _, ok := m.Primitives[handle]; ok {
		return handle
	}
	return m.RegisterMeshWithHandle(handle, cubeMesh(length))
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

func (m *ModelLibrary) getPrimitives(doc *modelspec.Document, node *modelspec.Node) []Primitive {
	q := []*modelspec.Node{node}

	var result []Primitive

	for len(q) > 0 {
		var nextLayerNodes []*modelspec.Node
		for _, node := range q {
			if node.MeshID != nil {
				mesh := doc.Meshes[*node.MeshID]

				modelConfig := &ModelConfig{MaxAnimationJointWeights: settings.MaxAnimationJointWeights}

				var vaos [][]uint32
				var geometryVAOs [][]uint32
				if m.processVisuals {
					vaos = createVAOs(modelConfig, []*modelspec.MeshSpecification{mesh})
					geometryVAOs = createGeometryVAOs(modelConfig, []*modelspec.MeshSpecification{mesh})
				}

				for i, primitive := range mesh.Primitives {
					p := Primitive{
						Primitive: primitive,
					}

					if m.processVisuals {
						p.VAO = vaos[0][i]
						p.GeometryVAO = geometryVAOs[0][i]
					}

					result = append(result, p)
				}
			}

			nextLayerNodes = append(nextLayerNodes, node.Children...)
		}
		q = nextLayerNodes
	}

	return result
}

func (m *ModelLibrary) RegisterSingleEntityDocument(document *modelspec.Document) {
	for _, scene := range document.Scenes {
		for _, node := range scene.Nodes {
			handle := NewGlobalHandle(document.Name)
			primitives := m.getPrimitives(document, node)
			m.Primitives[handle] = primitives
		}
	}
}

func (m *ModelLibrary) RegisterMesh(namespace string, mesh *modelspec.MeshSpecification) Handle {
	handle := NewHandleFromMeshID(namespace, mesh.ID)
	m.RegisterMeshWithHandle(handle, mesh)
	return handle
}

func (m *ModelLibrary) RegisterMeshWithHandle(handle Handle, mesh *modelspec.MeshSpecification) Handle {
	modelConfig := &ModelConfig{MaxAnimationJointWeights: settings.MaxAnimationJointWeights}

	var vaos [][]uint32
	var geometryVAOs [][]uint32
	if m.processVisuals {
		vaos = createVAOs(modelConfig, []*modelspec.MeshSpecification{mesh})
		geometryVAOs = createGeometryVAOs(modelConfig, []*modelspec.MeshSpecification{mesh})
	}

	for i, primitive := range mesh.Primitives {
		p := Primitive{
			Primitive: primitive,
		}

		if m.processVisuals {
			p.VAO = vaos[0][i]
			p.GeometryVAO = geometryVAOs[0][i]
		}

		m.Primitives[handle] = append(m.Primitives[handle], p)
	}
	return handle
}

func (m *ModelLibrary) RegisterAnimations(handle string, document *modelspec.Document) {
	m.Animations[handle] = document.Animations
	m.Joints[handle] = document.JointMap
	m.RootJoints[handle] = document.RootJoint.ID
}

func (m *ModelLibrary) GetAnimations(handle string) (map[string]*modelspec.AnimationSpec, map[int]*modelspec.JointSpec, int) {
	return m.Animations[handle], m.Joints[handle], m.RootJoints[handle]
}

func (m *ModelLibrary) GetPrimitives(handle Handle) []Primitive {
	if _, ok := m.Primitives[handle]; !ok {
		if handle.Type == HandleTypeCube {
			newHandle := m.GetOrCreateCubeMeshHandle(handle.IntTypeParam)
			handle = newHandle
		} else {
			return nil
		}
	}
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
