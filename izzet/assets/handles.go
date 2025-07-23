package assets

import (
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/kitolib/modelspec"
)

const (
	NamespaceGlobal  = "global"
	NamespaceDefault = "default"
)

var (
	DefaultMaterialHandle = types.MaterialHandle{Namespace: "global", ID: "0"}
	WhiteMaterialHandle   = types.MaterialHandle{Namespace: "global", ID: "1"}
)

type Primitive struct {
	Primitive *modelspec.PrimitiveSpecification

	// vao that contains all vertex attributes
	// position, normals, texture coords, joint indices/weights, etc
	VAO uint32

	// vao that only contains geometry related vertex attributes
	// i.e. vertex positions and joint indices / weights
	// but not normals, texture coords
	GeometryVAO uint32

	MaterialHandle types.MaterialHandle
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
func NewSingleEntityMeshHandle(namespace string) types.MeshHandle {
	return NewMeshHandle(namespace, "0")
}

func NewMeshHandle(namespace string, id string) types.MeshHandle {
	return types.MeshHandle{Namespace: namespace, ID: id}
}

func NewMaterialHandle(namespace string, id string) types.MaterialHandle {
	return types.MaterialHandle{Namespace: namespace, ID: id}
}

func (m *AssetManager) GetCubeMeshHandle() types.MeshHandle {
	return NewMeshHandle(NamespaceGlobal, "cube")
}

// this should probably look up a document, and get the animations from there, rather than storing these locally
func (m *AssetManager) GetAnimations(handle string) (map[string]*modelspec.AnimationSpec, map[int]*modelspec.JointSpec, int) {
	return m.Animations[handle], m.Joints[handle], m.RootJoints[handle]
}

func (m *AssetManager) GetPrimitives(handle types.MeshHandle) []Primitive {
	if _, ok := m.Primitives[handle]; !ok {
		return nil
	}
	return m.Primitives[handle]
}
