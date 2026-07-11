package entity

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/modelspec"
	"github.com/kkevinchou/izzet/izzet/assets"
)

var entityIDGen int

func InstantiateBaseEntity(name string, id int) *Entity {
	return &Entity{
		ID:   id,
		Name: name,

		Children: map[int]*Entity{},

		DirtyTransformFlag:   true,
		LocalPosition:        mgl64.Vec3{0, 0, 0},
		LocalRotation:        mgl64.QuatIdent(),
		LocalScale:           mgl64.Vec3{1, 1, 1},
		cachedWorldTransform: mgl64.Ident4(),
	}
}

func CreateEmptyEntity(name string) *Entity {
	entity := InstantiateBaseEntity(name, entityIDGen)
	entityIDGen += 1
	return entity
}

func SetNextID(nextID int) {
	entityIDGen = nextID
}

func GetNextIDAndAdvance() int {
	oldID := entityIDGen
	entityIDGen += 1
	return oldID
}

func BuildRelation(parent *Entity, child *Entity) {
	RemoveParent(child)
	parent.Children[child.ID] = child
	child.Parent = parent
	SetDirty(child)
}

func RemoveParent(child *Entity) {
	if parent := child.Parent; parent != nil {
		delete(parent.Children, child.ID)
		child.Parent = nil
	}
}

func AssetPrimitiveToSpecPrimitive(primitives []assets.Primitive) []*modelspec.Primitive {
	var result []*modelspec.Primitive
	for _, p := range primitives {
		result = append(result, p.Primitive)
	}
	return result
}

func VerticesFromNode(node *modelspec.Node, document *modelspec.Document, out *[]modelspec.Vertex) {
	if node.MeshID != nil {
		mesh := document.Meshes[*node.MeshID]
		for _, p := range mesh.Primitives {
			*out = append(*out, p.UniqueVertices...)
		}
	}

	for _, childNode := range node.Children {
		VerticesFromNode(childNode, document, out)
	}
}

func BatchRenderable(entity *Entity) bool {
	if !entity.Static || entity.MeshComponent == nil {
		return false
	}

	// default cubes are not supported for batch rendering since the texture
	// renders differently depending on the objects model space rotation and scale
	// we would need to expensively store this data for each mesh even if it
	// isn't a default cube

	isDefaultCubeWithDefaultMesh := entity.MeshComponent.MeshHandle == assets.DefaultCubeHandle &&
		len(entity.MeshComponent.Materials) > 0 &&
		entity.MeshComponent.Materials[0] == assets.DefaultMaterialID

	return !isDefaultCubeWithDefaultMesh
}

func GetPrimitiveMaterialIDs(am *assets.AssetManager, e *Entity) []assets.MaterialID {
	var materials []assets.MaterialID
	primitives := am.GetPrimitives(e.MeshComponent.MeshHandle)
	for i, prim := range primitives {
		// once we have prefabs working we should drop the use of materials from the primitive
		// the material from the primitive is the original material from the source asset
		materialID := prim.MaterialID
		if len(e.MeshComponent.Materials) > 0 && i < len(e.MeshComponent.Materials) {
			materialID = e.MeshComponent.Materials[i]
		}
		materials = append(materials, materialID)
	}
	return materials
}
