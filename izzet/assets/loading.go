package assets

import (
	"fmt"

	"github.com/kkevinchou/izzet/izzet/assets/loaders"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/kitolib/modelspec"
)

func (a *AssetManager) LoadAndRegisterDocument(name string, filepath string) bool {
	document := loaders.LoadDocument(name, filepath)
	if _, ok := a.documents[name]; ok {
		fmt.Printf("warning, document with name %s already previously loaded\n", name)
	}

	a.documents[name] = document
	a.RegisterSingleEntityDocument(document)
	return true
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
			m.registerMeshesInNode(document, node)
		}
	}
}

func (m *AssetManager) RegisterMesh(namespace string, mesh *modelspec.MeshSpecification) types.MeshHandle {
	handle := NewHandleFromMeshID(namespace, mesh.ID)
	m.registerMeshPrimitivesWithHandle(handle, mesh)
	return handle
}

func (m *AssetManager) RegisterAnimations(handle string, document *modelspec.Document) {
	m.Animations[handle] = document.Animations
	m.Joints[handle] = document.JointMap
	m.RootJoints[handle] = document.RootJoint.ID
}

func (m *AssetManager) registerMeshesInNode(doc *modelspec.Document, node *modelspec.Node) {
	handle := NewSingleMeshHandle(doc.Name)
	q := []*modelspec.Node{node}

	for len(q) > 0 {
		var nextLayerNodes []*modelspec.Node
		for _, node := range q {
			if node.MeshID != nil {
				mesh := doc.Meshes[*node.MeshID]
				m.registerMeshPrimitivesWithHandle(handle, mesh)
			}

			nextLayerNodes = append(nextLayerNodes, node.Children...)
		}
		q = nextLayerNodes
	}
}

func (m *AssetManager) registerMeshPrimitivesWithHandle(handle types.MeshHandle, mesh *modelspec.MeshSpecification) types.MeshHandle {
	var vaos [][]uint32
	var geometryVAOs [][]uint32
	if m.processVisuals {
		vaos = createVAOs([]*modelspec.MeshSpecification{mesh})
		geometryVAOs = createGeometryVAOs([]*modelspec.MeshSpecification{mesh})
	}

	for i, primitive := range mesh.Primitives {
		p := Primitive{
			Primitive: primitive,
		}

		if m.processVisuals {
			p.VAO = vaos[0][i]
			p.GeometryVAO = geometryVAOs[0][i]
			p.MaterialHandle = NewMaterialHandle(handle.Namespace, primitive.MaterialIndex)
		}

		m.Primitives[handle] = append(m.Primitives[handle], p)
	}
	return handle
}
