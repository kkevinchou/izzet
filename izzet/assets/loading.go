package assets

import (
	"fmt"

	"github.com/kkevinchou/izzet/izzet/assets/loaders"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/kitolib/modelspec"
)

func (a *AssetManager) LoadAndRegisterDocument(config AssetConfig) *modelspec.Document {
	document := loaders.LoadDocument(config.Name, config.FilePath)
	if _, ok := a.documentAssets[config.Name]; ok {
		fmt.Printf("document with name %s already previously loaded\n", config.Name)
	}

	a.documentAssets[config.Name] = DocumentAsset{
		Config:   config,
		Document: document,
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
	m.clearNamespace(document.Name)
	for _, mesh := range document.Meshes {
		m.registerMeshPrimitivesWithHandle(handle, mesh)
		m.NamespaceToMeshHandles[document.Name] = append(m.NamespaceToMeshHandles[document.Name], handle)
	}
}

func (m *AssetManager) registerDocumentMeshes(document *modelspec.Document) {
	m.clearNamespace(document.Name)
	for _, mesh := range document.Meshes {
		handle := NewMeshHandle(document.Name, fmt.Sprintf("%d", mesh.ID))
		m.registerMeshPrimitivesWithHandle(handle, mesh)
		m.NamespaceToMeshHandles[document.Name] = append(m.NamespaceToMeshHandles[document.Name], handle)
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

func (m *AssetManager) clearNamespace(namespace string) {
	for _, handle := range m.NamespaceToMeshHandles[namespace] {
		delete(m.Primitives, handle)
	}

	delete(m.NamespaceToMeshHandles, namespace)
}
