package entities

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/model"
)

type ShapeType string

var (
	CubeShapeType   ShapeType = "CUBE"
	SphereShapeType ShapeType = "SPHERE"
	LineShapeType   ShapeType = "LINE"
)

type CubeData struct {
	Length int
}

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
	Cube     *CubeData
	Sphere   *SphereData
	Capsule  *CapsuleData
	Line     *LineData
	Triangle *Triangle
}

// take an int so that we don't explode the number of VAOs we create
func CreateCube(length int) *Entity {
	entity := InstantiateBaseEntity("cube", id)
	entity.Model = model.NewCube()
	// entity.ShapeData = []*ShapeData{
	// 	&ShapeData{
	// 		Cube: &CubeData{
	// 			Length: length,
	// 		},
	// 	},
	// }
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
