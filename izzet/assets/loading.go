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
	a.RegisterDocumentMeshWithSingleHandle(document)
	return true
}

func (m *AssetManager) RegisterDocumentMeshWithSingleHandle(document *modelspec.Document) {
	handle := NewSingleMeshHandle(document.Name)
	for _, mesh := range document.Meshes {
		m.registerMeshPrimitivesWithHandle(handle, mesh)
	}
}

func (m *AssetManager) RegisterDocumentMeshes(document *modelspec.Document) {
	for _, mesh := range document.Meshes {
		handle := NewHandleFromMeshID(document.Name, mesh.ID)
		m.registerMeshPrimitivesWithHandle(handle, mesh)
	}
}

func (m *AssetManager) RegisterAnimations(handle string, document *modelspec.Document) {
	m.Animations[handle] = document.Animations
	m.Joints[handle] = document.JointMap
	m.RootJoints[handle] = document.RootJoint.ID
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
