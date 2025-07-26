package entities

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

func AssetPrimitiveToSpecPrimitive(primitives []assets.Primitive) []*modelspec.PrimitiveSpecification {
	var result []*modelspec.PrimitiveSpecification
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
