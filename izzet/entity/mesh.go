package entity

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/types"
)

type MeshComponent struct {
	MeshHandle             types.MeshHandle
	Transform              mgl64.Mat4
	Visible                bool
	InvisibleToPlayerOwner bool
	ShadowCasting          bool
}
