package entities

type ShapeType string

var (
	CubeShapeType   ShapeType = "CUBE"
	SphereShapeType ShapeType = "SPHERE"
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

type ShapeData struct {
	Type    ShapeType
	Cube    *CubeData
	Sphere  *SphereData
	Capsule *CapsuleData
}

func CreateCube() *Entity {
	entity := InstantiateBaseEntity("cube", id)
	entity.ShapeData = &ShapeData{Cube: &CubeData{Length: 50}}
	id += 1
	return entity
}
