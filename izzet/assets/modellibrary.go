package assets

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/kitolib/modelspec"
	"github.com/kkevinchou/kitolib/utils"
)

var nextGlobalID int

const (
	NamespaceGlobal = "global"
)

type ModelConfig struct {
	MaxAnimationJointWeights int
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

func NewGlobalHandle(id string) types.MeshHandle {
	return NewHandle(NamespaceGlobal, id)
}

func NewHandleFromMeshID(namespace string, meshID int) types.MeshHandle {
	return NewHandle(namespace, fmt.Sprintf("%d", meshID))
}

func NewHandle(namespace string, id string) types.MeshHandle {
	return types.MeshHandle{Namespace: namespace, ID: id}
}

func (m *AssetManager) GetCubeMeshHandle() types.MeshHandle {
	return NewHandle("global", "cube")
}

// TODO - need to answer questions around how we know what mesh data to reference when spawning an entity
//		- ideally we have a static and typed handle that we can easily reference from anywhere in the code
//		- this handle should be all we need to construct the mesh component
//		- the mesh component should be all we need to render entities in renderutils
//		- the handle should return all the primitives as well as the animations if any
//		- we need config to be able to mark a document as a single entity that's animated
//		- the registration API for AssetManager may need to be a whole document
//		- then the config determines what handle we want to associate with each asset
//			- Question, do I want to support selected instantiation of entities within a document?
//			- e.g. from within demo_scene_samurai, instantiating one entity by name

func (m *AssetManager) RegisterSingleEntityDocument(document *modelspec.Document) {
	for _, scene := range document.Scenes {
		for _, node := range scene.Nodes {
			handle := NewGlobalHandle(document.Name)
			primitives := m.getPrimitives(document, node)
			m.Primitives[handle] = primitives
		}
	}
}

func (m *AssetManager) RegisterMesh(namespace string, mesh *modelspec.MeshSpecification) types.MeshHandle {
	handle := NewHandleFromMeshID(namespace, mesh.ID)
	m.registerMeshWithHandle(handle, mesh)
	return handle
}

func (m *AssetManager) RegisterAnimations(handle string, document *modelspec.Document) {
	m.Animations[handle] = document.Animations
	m.Joints[handle] = document.JointMap
	m.RootJoints[handle] = document.RootJoint.ID
}

// this should probably look up a document, and get the animations from there, rather than storing these locally
func (m *AssetManager) GetAnimations(handle string) (map[string]*modelspec.AnimationSpec, map[int]*modelspec.JointSpec, int) {
	return m.Animations[handle], m.Joints[handle], m.RootJoints[handle]
}

func (m *AssetManager) GetPrimitives(handle types.MeshHandle) []Primitive {
	if _, ok := m.Primitives[handle]; !ok {
		return nil
	}
	return m.Primitives[handle]
}

func (m *AssetManager) getPrimitives(doc *modelspec.Document, node *modelspec.Node) []Primitive {
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

func (m *AssetManager) registerMeshWithHandle(handle types.MeshHandle, mesh *modelspec.MeshSpecification) types.MeshHandle {
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

// maybe this should be computed once and shared across all instances of the mesh?
func UniqueVerticesFromPrimitives(primitives []Primitive) []mgl64.Vec3 {
	var result []mgl64.Vec3
	for _, p := range primitives {
		result = append(result, utils.ModelSpecVertsToVec3(p.Primitive.UniqueVertices)...)
	}
	return result
}
