package assets

import (
	"fmt"

	"github.com/kkevinchou/izzet/izzet/assets/loaders"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/kitolib/modelspec"
)

func (a *AssetManager) LoadAndRegisterDocument(config AssetConfig) *modelspec.Document {
	document := loaders.LoadDocument(config.Name, config.FilePath)
	if _, ok := a.documents[config.Name]; ok {
		fmt.Printf("document with name %s already previously loaded\n", config.Name)
	}

	a.documents[config.Name] = Document{
		SourceFilePath: config.FilePath,
		Config:         config,
		Document:       document,
	}

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
		a.Animations[config.Name] = document.Animations
		a.Joints[config.Name] = document.JointMap
		a.RootJoints[config.Name] = document.RootJoint.ID
	}

	return document
}

func (m *AssetManager) registerDocumentMeshWithSingleHandle(document *modelspec.Document) {
	handle := NewSingleMeshHandle(document.Name)
	m.clearPrimitives(handle)
	for _, mesh := range document.Meshes {
		m.registerMeshPrimitivesWithHandle(handle, mesh)
	}
}

func (m *AssetManager) registerDocumentMeshes(document *modelspec.Document) {
	for _, mesh := range document.Meshes {
		handle := NewHandleFromMeshID(document.Name, mesh.ID)
		m.clearPrimitives(handle)
		m.registerMeshPrimitivesWithHandle(handle, mesh)
	}
}

func (m *AssetManager) registerAnimations(handle string, document *modelspec.Document) {
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

func (m *AssetManager) clearPrimitives(handle types.MeshHandle) {
	delete(m.Primitives, handle)
}
