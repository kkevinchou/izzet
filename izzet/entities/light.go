package entities

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/kitolib/utils"
)

type LightType int

const LightTypeDirection LightType = 0
const LightTypePoint LightType = 1

type LightInfo struct {
	Diffuse   mgl64.Vec4 // W component is the intensity
	Direction mgl64.Vec3
	Type      LightType
	Diffuse3F [3]float32
	Intensity float32
}

func (l *LightInfo) DiffuseVec4() mgl32.Vec4 {
	if l.Type == LightTypePoint {
		return mgl32.Vec4{l.Diffuse3F[0], l.Diffuse3F[1], l.Diffuse3F[2], float32(l.Intensity) * 100000}
	}
	return utils.Vec4F64ToF32(l.Diffuse)
}

// func (l *LightInfo) DiffuseVec3() mgl32.Vec3 {
// 	return utils.Vec4F64ToF32(l.Diffuse)
// }

func CreateLight(lightInfo *LightInfo) *Entity {
	entity := InstantiateBaseEntity("light", id)
	entity.ImageInfo = &ImageInfo{ImageName: "light.png"}
	entity.LightInfo = lightInfo
	entity.Billboard = &BillboardInfo{}
	SetScale(entity, mgl64.Vec3{10, 10, 10})
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
