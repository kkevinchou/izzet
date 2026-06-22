package assets

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/kkevinchou/izzet/internal/modelspec"
	"github.com/kkevinchou/izzet/izzet/assets/loaders"
)

func (a *AssetManager) LoadAndRegisterDocumentAsset(d Document) *modelspec.Document {
	start := time.Now()

	config := d.Config
	document := loaders.LoadDocument(config.Name, config.FilePath)
	if _, ok := a.documents[config.Name]; ok {
		fmt.Printf("document with name %s already previously loaded\n", config.Name)
	}

	a.clearDocumentPrimitives(config)
	d.Document = document
	a.documents[d.Config.Name] = d

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
	if _, ok := a.documents[config.Name]; ok {
		fmt.Printf("document with name %s already previously loaded\n", config.Name)
	}

	a.clearDocumentPrimitives(config)
	a.documents[config.Name] = Document{
		Config:        config,
		Document:      document,
		MatIDToHandle: map[string]MaterialHandle{},
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
			a.documents[config.Name].MatIDToHandle[material.ID] = handle
		}
	}

	a.registerDocumentMeshes(document, a.documents[config.Name].MatIDToHandle)

	if len(document.Animations) > 0 {
		a.Animations[config.Name] = document.Animations
		a.Joints[config.Name] = document.JointMap
		a.RootJoints[config.Name] = document.RootJoint.ID
	}

	a.logger.Info("LoadAndRegisterDocument", "name", document.Name, "time (ms)", time.Since(start).Milliseconds())

	return document
}

func (a *AssetManager) clearDocumentPrimitives(config AssetConfig) {
	delete(a.Primitives, newSingleEntityMeshHandle(config.Name))

	if existingAsset, ok := a.documents[config.Name]; ok && existingAsset.Document != nil {
		for _, mesh := range existingAsset.Document.Meshes {
			delete(a.Primitives, MeshHandle{namespace: config.Name, id: fmt.Sprintf("%d", mesh.ID)})
		}
	}
}

func createMaterialUniqueID(fp string, material modelspec.Material) string {
	split := strings.Split(filepath.ToSlash(fp), "/")
	return fmt.Sprintf("%s/%s", strings.Join(split[3:], "/"), material.ID)
}

func (m *AssetManager) registerDocumentMeshes(document *modelspec.Document, matIDToHandle map[string]MaterialHandle) {
	// registration of all primitives under one handle to support merged entity instantiation
	handle := newSingleEntityMeshHandle(document.Name)
	for _, mesh := range document.Meshes {
		m.registerMeshPrimitivesWithHandle(handle, mesh, matIDToHandle)
	}

	// per entity primitive registration
	for _, mesh := range document.Meshes {
		handle := MeshHandle{namespace: document.Name, id: fmt.Sprintf("%d", mesh.ID)}
		m.registerMeshPrimitivesWithHandle(handle, mesh, matIDToHandle)
	}
}

func (m *AssetManager) registerMeshPrimitivesWithHandle(handle MeshHandle, mesh *modelspec.Mesh, matIDToHandle map[string]MaterialHandle) MeshHandle {
	var vaos [][]uint32
	var geometryVAOs [][]uint32
	if m.processVisuals {
		vaos = createVAOs([]*modelspec.Mesh{mesh})
		geometryVAOs = createGeometryVAOs([]*modelspec.Mesh{mesh})
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
