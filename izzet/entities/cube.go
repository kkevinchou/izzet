package entities

import (
	"github.com/go-gl/mathgl/mgl64"
)

type ShapeType string

var (
	CubeShapeType   ShapeType = "CUBE"
	SphereShapeType ShapeType = "SPHERE"
	LineShapeType   ShapeType = "LINE"
)

type CubeData struct {
	Length float64
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

type ShapeData struct {
	Type    ShapeType
	Cube    *CubeData
	Sphere  *SphereData
	Capsule *CapsuleData
	Line    *LineData
}

func CreateCube() *Entity {
	entity := InstantiateBaseEntity("cube", id)
	entity.ShapeData = []*ShapeData{
		&ShapeData{
			Cube: &CubeData{
				50,
			},
		},
	}
	id += 1
	return entity
}
