package assets

import (
	"fmt"

	"github.com/kkevinchou/izzet/izzet/assets/loaders"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/kitolib/modelspec"
)

func (a *AssetManager) LoadAndRegisterDocument(config ImportAssetConfig) *modelspec.Document {
	document := loaders.LoadDocument(config.Name, config.FilePath)
	if _, ok := a.documents[config.Name]; ok {
		panic(fmt.Sprintf("document with name %s already previously loaded\n", config.Name))
	}

	a.documents[config.Name] = document

	if config.SingleEntity {
		a.registerDocumentMeshWithSingleHandle(document)
	} else {
		a.registerDocumentMeshes(document)
	}

	for _, material := range document.Materials {
		handle := NewMaterialHandle(document.Name, material.ID)
		a.Materials[handle] = material
	}

	if len(document.Animations) > 0 {
		a.registerAnimations(config.Name, document)
	}

	return document
}

func (m *AssetManager) registerDocumentMeshWithSingleHandle(document *modelspec.Document) {
	handle := NewSingleMeshHandle(document.Name)
	for _, mesh := range document.Meshes {
		m.registerMeshPrimitivesWithHandle(handle, mesh)
	}
}

func (m *AssetManager) registerDocumentMeshes(document *modelspec.Document) {
	for _, mesh := range document.Meshes {
		handle := NewHandleFromMeshID(document.Name, mesh.ID)
		m.registerMeshPrimitivesWithHandle(handle, mesh)
	}
}

func (m *AssetManager) registerAnimations(handle string, document *modelspec.Document) {
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
