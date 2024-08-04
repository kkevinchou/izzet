package entities

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/kkevinchou/izzet/izzet/types"
)

type MeshComponent struct {
	MeshHandle             types.MeshHandle
	Transform              mgl32.Mat4
	Visible                bool
	InvisibleToPlayerOwner bool
	ShadowCasting          bool
}
