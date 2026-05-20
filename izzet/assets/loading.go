package assets

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/kkevinchou/izzet/internal/modelspec"
	"github.com/kkevinchou/izzet/izzet/assets/loaders"
	"github.com/kkevinchou/izzet/izzet/types"
)

func (a *AssetManager) LoadAndRegisterDocumentAsset(d DocumentAsset) *modelspec.Document {
	start := time.Now()

	config := d.Config
	document := loaders.LoadDocument(config.Name, config.FilePath)
	if _, ok := a.documentAssets[config.Name]; ok {
		fmt.Printf("document with name %s already previously loaded\n", config.Name)
	}

	a.clearDocumentPrimitives(config)
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

	a.registerDocumentMeshes(document, d.MatIDToHandle)

	if len(document.Animations) > 0 {
		a.Animations[config.Name] = document.Animations
		a.Joints[config.Name] = document.JointMap
		a.RootJoints[config.Name] = document.RootJoint.ID
	}

	a.logger.Info("LoadAndRegisterDocumentAsset", "name", d.Document.Name, "time (ms)", time.Since(start).Milliseconds())

	return document
}

func (a *AssetManager) LoadAndRegisterDocument(config AssetConfig) *modelspec.Document {
	start := time.Now()

	document := loaders.LoadDocument(config.Name, config.FilePath)
	if _, ok := a.documentAssets[config.Name]; ok {
		fmt.Printf("document with name %s already previously loaded\n", config.Name)
	}

	a.clearDocumentPrimitives(config)
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

		for _, material := range document.Materials {
			name := fmt.Sprintf("%s/%s", document.Name, material.ID)
			handle := a.createMaterial(name, createMaterialUniqueID(config.FilePath, material), material)
			a.documentAssets[config.Name].MatIDToHandle[material.ID] = handle
		}
	}

	a.registerDocumentMeshes(document, a.documentAssets[config.Name].MatIDToHandle)

	if len(document.Animations) > 0 {
		a.Animations[config.Name] = document.Animations
		a.Joints[config.Name] = document.JointMap
		a.RootJoints[config.Name] = document.RootJoint.ID
	}

	a.logger.Info("LoadAndRegisterDocument", "name", document.Name, "time (ms)", time.Since(start).Milliseconds())

	return document
}

func (a *AssetManager) clearDocumentPrimitives(config AssetConfig) {
	delete(a.Primitives, NewSingleEntityMeshHandle(config.Name))

	if existingAsset, ok := a.documentAssets[config.Name]; ok && existingAsset.Document != nil {
		for _, mesh := range existingAsset.Document.Meshes {
			delete(a.Primitives, NewMeshHandle(config.Name, fmt.Sprintf("%d", mesh.ID)))
		}
	}
}

func createMaterialUniqueID(fp string, material modelspec.MaterialSpecification) string {
	split := strings.Split(filepath.ToSlash(fp), "/")
	return fmt.Sprintf("%s/%s", strings.Join(split[3:], "/"), material.ID)
}

func (m *AssetManager) registerDocumentMeshes(document *modelspec.Document, matIDToHandle map[string]types.MaterialHandle) {
	// registration of all primitives under one handle to support merged entity instantiation
	handle := NewSingleEntityMeshHandle(document.Name)
	for _, mesh := range document.Meshes {
		m.registerMeshPrimitivesWithHandle(handle, mesh, matIDToHandle)
	}

	// per entity primitive registration
	for _, mesh := range document.Meshes {
		handle := NewMeshHandle(document.Name, fmt.Sprintf("%d", mesh.ID))
		m.registerMeshPrimitivesWithHandle(handle, mesh, matIDToHandle)
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
				if _, ok := matIDToHandle[primitive.MaterialIndex]; !ok {
					panic("did not find material index in matIDToHandle map")
				}
				p.MaterialHandle = matIDToHandle[primitive.MaterialIndex]
			}
		}

		m.Primitives[handle] = append(m.Primitives[handle], p)
	}
	return handle
}
