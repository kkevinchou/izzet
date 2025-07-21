package collider

import (
	"github.com/go-gl/mathgl/mgl64"
)

var EmptyBoundingBox BoundingBox

type BoundingBox struct {
	MinVertex mgl64.Vec3
	MaxVertex mgl64.Vec3
}

func boundingBoxVertices(min, max mgl64.Vec3) []mgl64.Vec3 {
	delta := max.Sub(min)
	return []mgl64.Vec3{
		min,
		min.Add(mgl64.Vec3{delta[0], 0, 0}),
		min.Add(mgl64.Vec3{delta[0], delta[1], 0}),
		min.Add(mgl64.Vec3{0, delta[1], 0}),

		max,
		max.Sub(mgl64.Vec3{delta[0], 0, 0}),
		max.Sub(mgl64.Vec3{delta[0], delta[1], 0}),
		max.Sub(mgl64.Vec3{0, delta[1], 0}),
	}
}

func (c BoundingBox) Transform(mat mgl64.Mat4) BoundingBox {
	transformedVerts := boundingBoxVertices(c.MinVertex, c.MaxVertex)
	for i := range transformedVerts {
		transformedVerts[i] = mat.Mul4x1(transformedVerts[i].Vec4(1)).Vec3()
	}
	return BoundingBoxFromVertices(transformedVerts)
}

func BoundingBoxFromVertices(vertices []mgl64.Vec3) BoundingBox {
	var minX, minY, minZ, maxX, maxY, maxZ float64

	minX = vertices[0].X()
	maxX = vertices[0].X()
	minY = vertices[0].Y()
	maxY = vertices[0].Y()
	minZ = vertices[0].Z()
	maxZ = vertices[0].Z()
	for _, vertex := range vertices {
		if vertex.X() < minX {
			minX = vertex.X()
		}
		if vertex.X() > maxX {
			maxX = vertex.X()
		}
		if vertex.Y() < minY {
			minY = vertex.Y()
		}
		if vertex.Y() > maxY {
			maxY = vertex.Y()
		}
		if vertex.Z() < minZ {
			minZ = vertex.Z()
		}
		if vertex.Z() > maxZ {
			maxZ = vertex.Z()
		}
	}

	minVertex := mgl64.Vec3{minX, minY, minZ}
	maxVertex := mgl64.Vec3{maxX, maxY, maxZ}

	return BoundingBox{MinVertex: minVertex, MaxVertex: maxVertex}
}

func BoundingBoxFromCapsule(capsule Capsule) *BoundingBox {
	return &BoundingBox{
		MinVertex: capsule.Bottom.Add(mgl64.Vec3{-capsule.Radius, -capsule.Radius, -capsule.Radius}),
		MaxVertex: capsule.Top.Add(mgl64.Vec3{capsule.Radius, capsule.Radius, capsule.Radius}),
	}
}

func NewBoundingBox(min, max mgl64.Vec3) *BoundingBox {
	return &BoundingBox{
		MinVertex: min,
		MaxVertex: max,
	}
}
