package entities

import (
	"github.com/go-gl/mathgl/mgl32"
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
	Radius float32
}

type CapsuleData struct {
	Radius float32
	Length float32
}

type LineData struct {
	Vector mgl32.Vec3
}

type Triangle struct {
	V1 mgl32.Vec3
	V2 mgl32.Vec3
	V3 mgl32.Vec3
}

type ShapeData struct {
	Type     ShapeType
	Sphere   *SphereData
	Capsule  *CapsuleData
	Line     *LineData
	Triangle *Triangle
}

func CreateCube(ml *assets.AssetManager, length float32) *Entity {
	entity := InstantiateBaseEntity("cube", id)
	entity.LocalScale = mgl32.Vec3{length, length, length}

	handle := ml.GetCubeMeshHandle()
	entity.MeshComponent = &MeshComponent{
		MeshHandle:    handle,
		Transform:     mgl32.Ident4(),
		Visible:       true,
		ShadowCasting: true,
	}

	// cube only has a singular primitives
	primitives := ml.GetPrimitives(handle)
	uniqueVertices := utils.ModelSpecVertsToVec3(primitives[0].Primitive.UniqueVertices)
	entity.InternalBoundingBox = collider.BoundingBoxFromVertices(uniqueVertices)

	rotation := mgl32.QuatRotate(90, mgl32.Vec3{1, 0, 0})
	rotation = rotation.Mul(mgl32.QuatRotate(90, mgl32.Vec3{0, 0, -1}))
	entity.Physics = &PhysicsComponent{Velocity: mgl32.Vec3{0, 0, 0}}

	id += 1
	return entity
}

func CreateTriangle(v1, v2, v3 mgl32.Vec3) *Entity {
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
