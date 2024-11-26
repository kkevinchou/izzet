package assets

import (
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/kitolib/modelspec"
)

const (
	NamespaceGlobal = "global"
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

func NewSingleMeshHandle(namespace string) types.MeshHandle {
	return NewMeshHandle(namespace, "0")
}

func NewMeshHandle(namespace string, id string) types.MeshHandle {
	return types.MeshHandle{Namespace: namespace, ID: id}
}

func NewMaterialHandle(namespace string, id string) types.MaterialHandle {
	return types.MaterialHandle{Namespace: namespace, ID: id}
}

func (m *AssetManager) GetCubeMeshHandle() types.MeshHandle {
	return NewMeshHandle("global", "cube")
}

func (m *AssetManager) GetDefaultMaterialHandle() types.MaterialHandle {
	return types.MaterialHandle{Namespace: "global"}
}

func (m *AssetManager) GetMaterial(handle types.MaterialHandle) modelspec.MaterialSpecification {
	if material, ok := m.Materials[handle]; ok {
		return material
	}
	return m.Materials[m.GetDefaultMaterialHandle()]
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
