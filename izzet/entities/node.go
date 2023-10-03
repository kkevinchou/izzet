package entities

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/kkevinchou/izzet/izzet/modellibrary"
)

type Node struct {
	// todo - remove this and use mesh handles
	MeshID     int
	MeshHandle modellibrary.Handle
	Transform  mgl32.Mat4
	Children   []Node
}
