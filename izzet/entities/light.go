package entities

import "github.com/go-gl/mathgl/mgl64"

type LightInfo struct {
	Diffuse   mgl64.Vec3
	Direction mgl64.Vec3
	Type      int // 0 - directional
}

func CreateLight(lightInfo *LightInfo) *Entity {
	entity := InstantiateBaseEntity("light", id)
	entity.LightInfo = lightInfo
	id += 1
	return entity
}
