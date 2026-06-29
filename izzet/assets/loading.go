package assets

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/kkevinchou/izzet/internal/modelspec"
	"github.com/kkevinchou/izzet/izzet/assets/loaders"
)

func (a *AssetManager) ReloadDocument(d Document) *modelspec.Document {
	start := time.Now()

	document := loaders.LoadDocument(d.ID, d.Filepath)
	if _, ok := a.documents[d.ID]; ok {
		fmt.Printf("document with name %s already previously loaded\n", d.ID)
	}

	a.clearDocumentPrimitives(d.ID)
	d.Document = document
	a.documents[d.ID] = d

	if a.processVisuals {
		for _, file := range document.PeripheralFiles {
			extension := filepath.Ext(file)
			if extension != ".png" {
				continue
			}

			key := file[0 : len(file)-len(extension)]
			a.textures[key] = loaders.LoadTexture(filepath.Join(filepath.Dir(d.Filepath), file))
		}
	}

	a.registerDocumentMeshes(document, d.SourceMaterialIndexToMaterialID)

	if len(document.Animations) > 0 {
		a.Animations[d.ID] = document.Animations
		a.Joints[d.ID] = document.JointMap
		a.RootJoints[d.ID] = document.RootJoint.ID
	}

	a.logger.Info("ReloadDocument", "name", d.Document.Name, "time (ms)", time.Since(start).Milliseconds())

	return document
}

func (a *AssetManager) ImportDocument(id string, path string) *modelspec.Document {
	start := time.Now()

	document := loaders.LoadDocument(id, path)
	if _, ok := a.documents[id]; ok {
		fmt.Printf("document with name %s already previously loaded\n", id)
	}

	a.clearDocumentPrimitives(id)
	sourceMaterialIndexToMaterialID := createSourceMaterialIndexMap(path, document)
	a.documents[id] = Document{
		ID:                              id,
		Filepath:                        path,
		Document:                        document,
		SourceMaterialIndexToMaterialID: sourceMaterialIndexToMaterialID,
	}

	if a.processVisuals {
		for _, file := range document.PeripheralFiles {
			extension := filepath.Ext(file)
			if extension != ".png" {
				continue
			}

			key := file[0 : len(file)-len(extension)]
			a.textures[key] = loaders.LoadTexture(filepath.Join(filepath.Dir(path), file))
		}

		for i, material := range document.Materials {
			name := fmt.Sprintf("%s/%s", document.Name, material.ID)
			a.createMaterial(name, sourceMaterialIndexToMaterialID[i], material)
		}
	}

	a.registerDocumentMeshes(document, a.documents[id].SourceMaterialIndexToMaterialID)

	if len(document.Animations) > 0 {
		a.Animations[id] = document.Animations
		a.Joints[id] = document.JointMap
		a.RootJoints[id] = document.RootJoint.ID
	}

	a.logger.Info("LoadAndRegisterDocument", "name", document.Name, "time (ms)", time.Since(start).Milliseconds())

	return document
}

func (a *AssetManager) clearDocumentPrimitives(name string) {
	delete(a.Primitives, newSingleEntityMeshHandle(name))

	if existingAsset, ok := a.documents[name]; ok && existingAsset.Document != nil {
		for _, mesh := range existingAsset.Document.Meshes {
			delete(a.Primitives, MeshHandle{namespace: name, id: fmt.Sprintf("%d", mesh.ID)})
		}
	}
}

// this material ID needs to be deterministic so that we don't recreate the same materials
// over and over when we reload the gltf file
func createStableMaterialID(fp string, material modelspec.Material) MaterialID {
	split := strings.Split(filepath.ToSlash(fp), "/")
	return MaterialID(fmt.Sprintf("%s/%s", strings.Join(split[3:], "/"), material.ID))
}

func createSourceMaterialIndexMap(fp string, document *modelspec.Document) map[int]MaterialID {
	sourceMaterialIndexToMaterialID := map[int]MaterialID{}
	for i, material := range document.Materials {
		sourceMaterialIndexToMaterialID[i] = createStableMaterialID(fp, material)
	}
	return sourceMaterialIndexToMaterialID
}

func (m *AssetManager) registerDocumentMeshes(document *modelspec.Document, sourceMaterialIndexToMaterialID map[int]MaterialID) {
	// registration of all primitives under one handle to support merged entity instantiation
	handle := newSingleEntityMeshHandle(document.Name)
	for _, mesh := range document.Meshes {
		m.registerMeshPrimitivesWithHandle(handle, mesh, sourceMaterialIndexToMaterialID)
	}

	// per entity primitive registration
	for _, mesh := range document.Meshes {
		handle := MeshHandle{namespace: document.Name, id: fmt.Sprintf("%d", mesh.ID)}
		m.registerMeshPrimitivesWithHandle(handle, mesh, sourceMaterialIndexToMaterialID)
	}
}

func (m *AssetManager) registerMeshPrimitivesWithHandle(handle MeshHandle, mesh *modelspec.Mesh, sourceMaterialIndexToMaterialID map[int]MaterialID) MeshHandle {
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

		if primitive.MaterialIndex != nil {
			materialID, ok := sourceMaterialIndexToMaterialID[*primitive.MaterialIndex]
			if !ok {
				panic("did not find material index in sourceMaterialIndexToMaterialID map")
			}
			p.MaterialID = materialID
		}

		if m.processVisuals {
			p.VAO = vaos[0][i]
			p.GeometryVAO = geometryVAOs[0][i]
		}

		m.Primitives[handle] = append(m.Primitives[handle], p)
	}
	return handle
}
