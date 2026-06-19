package entity

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/assets/handle"
)

type MeshComponent struct {
	MeshHandle             handle.Mesh
	Transform              mgl64.Mat4
	Visible                bool
	InvisibleToPlayerOwner bool
	ShadowCasting          bool
}
