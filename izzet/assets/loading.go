package assets

import (
	"fmt"
	"path/filepath"

	"github.com/kkevinchou/izzet/izzet/assets/loaders"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/kitolib/modelspec"
)

func (a *AssetManager) LoadAndRegisterDocumentAsset(d DocumentAsset) *modelspec.Document {
	config := d.Config
	document := loaders.LoadDocument(config.Name, config.FilePath)
	if _, ok := a.documentAssets[config.Name]; ok {
		fmt.Printf("document with name %s already previously loaded\n", config.Name)
	}

	d.Document = document
	a.documentAssets[d.Config.Name] = d

	if a.processVisuals {
		for _, file := range document.PeripheralFiles {
			extension := filepath.Ext(file)
			if extension != ".png" {
				continue
			}

			key := file[0 : len(file)-len(extension)]
			a.textures[key] = loaders.LoadTexture(filepath.Join(filepath.Dir(config.FilePath), file))
		}
	}

	if config.SingleEntity {
		// look up the material -> handle setup on the load call
		// materials have been saved to the assets json and should be initialized
		// in initializeAssetManagerWithProject()
		// 1. LoadAndRegisterDocument
		// 2. CreateMaterialWithHandle for each material
		//
		// document should contain a mapping from the gltf material -> a material handle
		a.registerDocumentMeshWithSingleHandle(document, d.MatIDToHandle)
	} else {
		a.registerDocumentMeshes(document, d.MatIDToHandle)
	}

	if len(document.Animations) > 0 {
		a.Animations[config.Name] = document.Animations
		a.Joints[config.Name] = document.JointMap
		a.RootJoints[config.Name] = document.RootJoint.ID
	}

	return document
}

func (a *AssetManager) LoadAndRegisterDocument(config AssetConfig, importMaterials bool) *modelspec.Document {
	document := loaders.LoadDocument(config.Name, config.FilePath)
	if _, ok := a.documentAssets[config.Name]; ok {
		fmt.Printf("document with name %s already previously loaded\n", config.Name)
	}

	a.documentAssets[config.Name] = DocumentAsset{
		Config:        config,
		Document:      document,
		MatIDToHandle: map[string]types.MaterialHandle{},
	}

	if a.processVisuals {
		for _, file := range document.PeripheralFiles {
			extension := filepath.Ext(file)
			if extension != ".png" {
				continue
			}

			key := file[0 : len(file)-len(extension)]
			a.textures[key] = loaders.LoadTexture(filepath.Join(filepath.Dir(config.FilePath), file))
		}

		if importMaterials {
			for _, material := range document.Materials {
				name := fmt.Sprintf("%s/%s", document.Name, material.ID)
				handle := a.createMaterial(name, fmt.Sprintf("%s/%s", config.FilePath, material.ID), material)
				a.documentAssets[config.Name].MatIDToHandle[material.ID] = handle
			}
		}
	}

	if config.SingleEntity {
		// look up the material -> handle setup on the load call
		// materials have been saved to the assets json and should be initialized
		// in initializeAssetManagerWithProject()
		// 1. LoadAndRegisterDocument
		// 2. CreateMaterialWithHandle for each material
		//
		// document should contain a mapping from the gltf material -> a material handle
		a.registerDocumentMeshWithSingleHandle(document, a.documentAssets[config.Name].MatIDToHandle)
	} else {
		a.registerDocumentMeshes(document, a.documentAssets[config.Name].MatIDToHandle)
	}

	if len(document.Animations) > 0 {
		a.Animations[config.Name] = document.Animations
		a.Joints[config.Name] = document.JointMap
		a.RootJoints[config.Name] = document.RootJoint.ID
	}

	return document
}

func (m *AssetManager) registerDocumentMeshWithSingleHandle(document *modelspec.Document, matIDToHandle map[string]types.MaterialHandle) {
	handle := NewSingleEntityMeshHandle(document.Name)
	m.clearNamespace(document.Name)
	for _, mesh := range document.Meshes {
		m.registerMeshPrimitivesWithHandle(handle, mesh, matIDToHandle)
		m.NamespaceToMeshHandles[document.Name] = append(m.NamespaceToMeshHandles[document.Name], handle)
	}
}

func (m *AssetManager) registerDocumentMeshes(document *modelspec.Document, matIDToHandle map[string]types.MaterialHandle) {
	m.clearNamespace(document.Name)
	for _, mesh := range document.Meshes {
		handle := NewMeshHandle(document.Name, fmt.Sprintf("%d", mesh.ID))
		m.registerMeshPrimitivesWithHandle(handle, mesh, matIDToHandle)
		m.NamespaceToMeshHandles[document.Name] = append(m.NamespaceToMeshHandles[document.Name], handle)
	}
}

func (m *AssetManager) registerMeshPrimitivesWithHandle(handle types.MeshHandle, mesh *modelspec.MeshSpecification, matIDToHandle map[string]types.MaterialHandle) types.MeshHandle {
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
			if len(matIDToHandle) > 0 {
				p.MaterialHandle = matIDToHandle[primitive.MaterialIndex]
			}
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
