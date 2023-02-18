package entities

type ShapeType string

var (
	CubeShapeType   ShapeType = "CUBE"
	SphereShapeType ShapeType = "SPHERE"
)

type CubeData struct {
	Width  float64
	Length float64
	Height float64
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
	entity.ShapeData = &ShapeData{Cube: &CubeData{Width: 25, Length: 25, Height: 25}}
	id += 1
	return entity
}
