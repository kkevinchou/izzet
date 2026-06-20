package assets

import (
	"encoding/json"
	"strings"

	"github.com/kkevinchou/izzet/internal/modelspec"
)

var (
	defaultMaterialHandle = newMaterialHandle("custom/default")
	whiteMaterialHandle   = newMaterialHandle("custom/white")
	defaultCubeHandle     = newMeshHandle("global", "cube")
)

type opaqueHandleValue string

func (v opaqueHandleValue) String() string {
	return "<opaque-handle-value>"
}

func (v opaqueHandleValue) GoString() string {
	return v.String()
}

type AnimationHandle struct {
	id opaqueHandleValue
}

func newAnimationHandle(id string) AnimationHandle {
	return AnimationHandle{id: opaqueHandleValue(id)}
}

func (h AnimationHandle) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(h.id))
}

func (h *AnimationHandle) UnmarshalJSON(data []byte) error {
	var id string
	if err := json.Unmarshal(data, &id); err != nil {
		return err
	}
	*h = newAnimationHandle(id)
	return nil
}

type MeshHandle struct {
	namespace opaqueHandleValue
	id        opaqueHandleValue
}

func newMeshHandle(namespace string, id string) MeshHandle {
	return MeshHandle{namespace: opaqueHandleValue(namespace), id: opaqueHandleValue(id)}
}

func (h MeshHandle) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Namespace string
		ID        string
	}{
		Namespace: string(h.namespace),
		ID:        string(h.id),
	})
}

func (h *MeshHandle) UnmarshalJSON(data []byte) error {
	var value struct {
		Namespace string
		ID        string
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*h = newMeshHandle(value.Namespace, value.ID)
	return nil
}

type MaterialHandle struct {
	id opaqueHandleValue
}

func newMaterialHandle(id string) MaterialHandle {
	return MaterialHandle{id: opaqueHandleValue(id)}
}

func (h MaterialHandle) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID string
	}{
		ID: string(h.id),
	})
}

func (h *MaterialHandle) UnmarshalJSON(data []byte) error {
	var value struct {
		ID string
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*h = newMaterialHandle(value.ID)
	return nil
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

	MaterialHandle MaterialHandle
}

// CONTEXT
// - an entity should only ever have 1 mesh handle, which can point to many primitives for rendering
// - when we want to create an entity from a document, we should be able to convert the document name to the mesh handle
//   - creating an entity in this way only instantiates some baseline entity properties e.g. transforms, colliders
//   - other parts of the entity may still need to be setup

// REFACTOR
// - ideally we deprecate the use of handles as a combination of namespace + id. it should just be a singular id
// - when we import a collection of meshes/materials from a document, each mesh/material should be treated like its own asset
//   and can be instantiated as such individually. right now we put everything under a singular document and it is all
//   instantiated together
// - documents are one:many meshes/materials

// FUTURE WORK
// - importing a document right now is doing two things:
//   - setting up a repeatable way to instantiate the document (a form of prefab)
//   - extracting meshes/materials/animations into the content browser
//   - ideally we can treat this as two separate operations so that we have the ability to instantiate the entire level
//     or just instantiatign individual meshes from the document
//
// - the instantiated entities should have meshes/materials that point to loaded meshes/materials
func newSingleEntityMeshHandle(namespace string) MeshHandle {
	return newMeshHandle(namespace, "__merged__")
}

func (m *AssetManager) DefaultMaterialHandle() MaterialHandle {
	return defaultMaterialHandle
}

func (m *AssetManager) WhiteMaterialHandle() MaterialHandle {
	return whiteMaterialHandle
}

func (m *AssetManager) DefaultCubeHandle() MeshHandle {
	return defaultCubeHandle
}

func (m *AssetManager) GetSingleEntityMeshHandle(namespace string) MeshHandle {
	return newSingleEntityMeshHandle(namespace)
}

func (m *AssetManager) GetDocumentMeshHandle(namespace string, id string) MeshHandle {
	return newMeshHandle(namespace, id)
}

func (m *AssetManager) MeshHandleReferencesDocument(meshHandle MeshHandle, namespace string) bool {
	return string(meshHandle.namespace) == namespace
}

func (m *AssetManager) GetAnimationHandle(id string) AnimationHandle {
	return newAnimationHandle(id)
}

func (m *AssetManager) IsRaptorAnimationHandle(animationHandle AnimationHandle) bool {
	return strings.Contains(string(animationHandle.id), "velociraptor")
}

// this should probably look up a document, and get the animations from there, rather than storing these locally
func (m *AssetManager) GetAnimations(animationHandle AnimationHandle) (map[string]*modelspec.AnimationSpec, map[int]*modelspec.JointSpec, int) {
	id := string(animationHandle.id)
	return m.Animations[id], m.Joints[id], m.RootJoints[id]
}

func (m *AssetManager) GetPrimitives(meshHandle MeshHandle) []Primitive {
	if _, ok := m.Primitives[meshHandle]; !ok {
		return nil
	}
	return m.Primitives[meshHandle]
}
