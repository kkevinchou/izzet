package entities

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/kkevinchou/izzet/izzet/modellibrary"
)

type MeshComponent struct {
	Node Node
}

type Node struct {
	MeshHandle *modellibrary.Handle
	Transform  mgl32.Mat4
	Children   []Node
}
