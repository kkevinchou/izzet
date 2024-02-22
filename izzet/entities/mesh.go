package entities

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/modellibrary"
)

type MeshComponent struct {
	MeshHandle             modellibrary.MeshHandle
	Transform              mgl64.Mat4
	Visible                bool
	InvisibleToPlayerOwner bool
	ShadowCasting          bool
}
