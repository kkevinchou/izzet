package entities

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/modellibrary"
)

type MeshComponent struct {
	MeshHandle modellibrary.Handle
	Transform  mgl64.Mat4
}
