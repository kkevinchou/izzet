package entities

import "github.com/go-gl/mathgl/mgl64"

type LightInfo struct {
	Diffuse   mgl64.Vec4 // W component is the intensity
	Direction mgl64.Vec3
	Type      int // 0 - directional
}

func CreateLight(lightInfo *LightInfo) *Entity {
	entity := InstantiateBaseEntity("light", id)
	entity.ImageInfo = &ImageInfo{ImageName: "light.png"}
	entity.LightInfo = lightInfo
	entity.Scale = mgl64.Vec3{10, 10, 10}
	// entity.ShapeData = []*ShapeData{
	// 	&ShapeData{
	// 		Cube: &CubeData{
	// 			15,
	// 		},
	// 	},
	// }
	id += 1
	return entity
}