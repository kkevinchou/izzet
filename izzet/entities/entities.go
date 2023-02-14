package entities

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/kkevinchou/kitolib/animation"
	"github.com/kkevinchou/kitolib/modelspec"
)

var id int

type Entity struct {
	ID   int
	Name string

	Parent *Entity

	// each Entity has their own transforms and animation player
	Position mgl64.Vec3
	Rotation mgl64.Quat
	Scale    mgl64.Vec3

	// native objects
	// -- cube, capsule, cylinder, etc

	// 3D imported models
	Prefab          *prefabs.Prefab
	Animations      map[string]*modelspec.AnimationSpec
	AnimationPlayer *animation.AnimationPlayer
}

func SetNextID(nextID int) {
	id = nextID
}

func InstantiateFromPrefab(prefab *prefabs.Prefab) *Entity {
	e := InstantiateFromPrefabStaticID(id, prefab)
	id += 1
	return e
}

func InstantiateFromPrefabStaticID(id int, prefab *prefabs.Prefab) *Entity {
	e := &Entity{
		ID:   id,
		Name: fmt.Sprintf("%s-%d", prefab.Name, id),

		Position: mgl64.Vec3{0, 0, 0},
		Rotation: mgl64.QuatIdent(),
		Scale:    mgl64.Vec3{1, 1, 1},

		Prefab: prefab,
	}

	// animation setup
	e.Animations = prefab.ModelRefs[0].Model.Animations()
	if len(e.Animations) > 0 {
		e.AnimationPlayer = animation.NewAnimationPlayer(prefab.ModelRefs[0].Model)
	}

	return e
}

// ComputeTransformMatrix calculates the final transform matrix for model
// by traversing its parental hierarchy if it exists
func ComputeTransformMatrix(entity *Entity) mgl64.Mat4 {
	translationMatrix := mgl64.Translate3D(entity.Position[0], entity.Position[1], entity.Position[2])
	rotationMatrix := entity.Rotation.Mat4()
	scaleMatrix := mgl64.Scale3D(entity.Scale.X(), entity.Scale.Y(), entity.Scale.Z())

	modelMatrix := translationMatrix.Mul4(rotationMatrix).Mul4(scaleMatrix)
	if entity.Parent != nil {
		parentModelMatrix := ComputeTransformMatrix(entity.Parent)
		modelMatrix = parentModelMatrix.Mul4(modelMatrix)
	}

	return modelMatrix
}
