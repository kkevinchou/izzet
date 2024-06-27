package entities

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/kitolib/collision/collider"
	"github.com/kkevinchou/kitolib/utils"
)

type ShapeType string

var (
	CubeShapeType   ShapeType = "CUBE"
	SphereShapeType ShapeType = "SPHERE"
	LineShapeType   ShapeType = "LINE"
)

type SphereData struct {
	Radius float64
}

type CapsuleData struct {
	Radius float64
	Length float64
}

type LineData struct {
	Vector mgl64.Vec3
}

type Triangle struct {
	V1 mgl64.Vec3
	V2 mgl64.Vec3
	V3 mgl64.Vec3
}

type ShapeData struct {
	Type     ShapeType
	Sphere   *SphereData
	Capsule  *CapsuleData
	Line     *LineData
	Triangle *Triangle
}

func CreateCube(ml *assets.AssetManager, length float64) *Entity {
	entity := InstantiateBaseEntity("cube", id)
	entity.LocalScale = mgl64.Vec3{length, length, length}

	handle := ml.GetCubeMeshHandle()
	entity.MeshComponent = &MeshComponent{
		MeshHandle:    handle,
		Transform:     mgl64.Ident4(),
		Visible:       true,
		ShadowCasting: true,
	}

	// cube only has a singular primitives
	primitives := ml.GetPrimitives(handle)
	uniqueVertices := utils.ModelSpecVertsToVec3(primitives[0].Primitive.UniqueVertices)
	entity.InternalBoundingBox = collider.BoundingBoxFromVertices(uniqueVertices)

	rotation := mgl64.QuatRotate(90, mgl64.Vec3{1, 0, 0})
	rotation = rotation.Mul(mgl64.QuatRotate(90, mgl64.Vec3{0, 0, -1}))
	entity.Physics = &PhysicsComponent{Velocity: mgl64.Vec3{0, 0, 0}}

	id += 1
	return entity
}

func CreateTriangle(v1, v2, v3 mgl64.Vec3) *Entity {
	entity := InstantiateBaseEntity("triangle", id)
	entity.ShapeData = []*ShapeData{
		&ShapeData{
			Triangle: &Triangle{
				V1: v1,
				V2: v2,
				V3: v3,
			},
		},
	}
	id += 1
	return entity
}
