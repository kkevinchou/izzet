package entity

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/assets"
)

type MeshComponent struct {
	MeshHandle             assets.MeshHandle
	Materials              []assets.MaterialHandle
	Transform              mgl64.Mat4
	Visible                bool
	InvisibleToPlayerOwner bool
	ShadowCasting          bool
}
