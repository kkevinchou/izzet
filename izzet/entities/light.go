package entities

import (
	"github.com/go-gl/mathgl/mgl32"
)

type LightType int

const LightTypeDirection LightType = 0
const LightTypePoint LightType = 1

type LightInfo struct {
	Range              float32
	Direction3F        [3]float32
	Type               LightType
	Diffuse3F          [3]float32
	PreScaledIntensity float32
}

func (l *LightInfo) IntensifiedDiffuse() mgl32.Vec3 {
	return mgl32.Vec3{l.Diffuse3F[0], l.Diffuse3F[1], l.Diffuse3F[2]}.Mul(l.Intensity())
}

// we scale the intensity value for point lights so that it's more user friendly to manage
// the sliders in the UI are in a small range (< 100) rather than in the hundreds of thousands.
// this is still a little confusing so I'll probably revisit this at some point
func (l *LightInfo) Intensity() float32 {
	intensityScale := 1
	if l.Type == LightTypePoint {
		intensityScale = 100000
	}
	return l.PreScaledIntensity * float32(intensityScale)
}

func CreateDirectionalLight() *Entity {
	lightInfo := &LightInfo{
		PreScaledIntensity: 2,
		Diffuse3F:          [3]float32{1, 1, 1},
		Type:               LightTypeDirection,
		Direction3F:        [3]float32{-1, -1, -1},
	}

	entity := InstantiateBaseEntity("directional-light", id)
	entity.ImageInfo = NewImageInfo("lamp.png", 5)
	entity.LightInfo = lightInfo
	entity.Billboard = true
	id += 1
	return entity
}

func CreatePointLight() *Entity {
	lightInfo := &LightInfo{
		PreScaledIntensity: 0.1,
		Diffuse3F:          [3]float32{1, 1, 1},
		Type:               LightTypePoint,
		Range:              800,
	}

	entity := InstantiateBaseEntity("point-light", id)
	entity.ImageInfo = NewImageInfo("lamp.png", 5)
	entity.LightInfo = lightInfo
	entity.Billboard = true
	id += 1
	return entity
}
