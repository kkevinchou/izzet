package collider

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/lib/libutils"
	"github.com/kkevinchou/izzet/lib/model"
)

type Mesh interface {
	Vertices() []mgl64.Vec3
}

type Triangle struct {
	Normal mgl64.Vec3
	Points []mgl64.Vec3
}

func (t Triangle) Transform(transform mgl64.Mat4) Triangle {
	return NewTriangle([]mgl64.Vec3{
		transform.Mul4x1(t.Points[0].Vec4(1)).Vec3(),
		transform.Mul4x1(t.Points[1].Vec4(1)).Vec3(),
		transform.Mul4x1(t.Points[2].Vec4(1)).Vec3(),
	})
}

func NewTriangle(points []mgl64.Vec3) Triangle {
	seg1 := points[1].Sub(points[0])
	seg2 := points[2].Sub(points[0])
	normal := seg1.Cross(seg2).Normalize()
	return Triangle{
		Points: points,
		Normal: normal,
	}
}

type TriMesh struct {
	Triangles []Triangle
}

func (t TriMesh) Transform(transform mgl64.Mat4) TriMesh {
	newTriMesh := TriMesh{Triangles: make([]Triangle, len(t.Triangles))}
	for i, tri := range t.Triangles {
		newTriMesh.Triangles[i] = tri.Transform(transform)
	}
	return newTriMesh
}

func NewTriMesh(model *model.Model) TriMesh {
	triMesh := TriMesh{}
	for _, mesh := range model.Meshes() {
		for _, meshChunk := range mesh.MeshChunks() {
			vertices := meshChunk.Vertices()
			for i := 0; i < len(vertices); i += 3 {
				points := []mgl64.Vec3{
					libutils.Vec3F32ToF64(vertices[i].Position),
					libutils.Vec3F32ToF64(vertices[i+1].Position),
					libutils.Vec3F32ToF64(vertices[i+2].Position),
				}
				triMesh.Triangles = append(triMesh.Triangles, NewTriangle(points))
			}
		}
	}
	return triMesh
}

func NewBoxTriMesh(w, l, h float64) TriMesh {
	halfW := w / 2
	halfL := l / 2
	triMesh := TriMesh{}
	// front
	triMesh.Triangles = append(triMesh.Triangles, NewTriangle(
		[]mgl64.Vec3{{-halfW, 0, halfL}, {halfW, 0, halfL}, {halfW, h, halfL}},
	))
	triMesh.Triangles = append(triMesh.Triangles, NewTriangle(
		[]mgl64.Vec3{{halfW, h, halfL}, {-halfW, h, halfL}, {-halfW, 0, halfL}},
	))
	// back
	triMesh.Triangles = append(triMesh.Triangles, NewTriangle(
		[]mgl64.Vec3{{halfW, 0, -halfL}, {-halfW, 0, -halfL}, {-halfW, h, -halfL}},
	))
	triMesh.Triangles = append(triMesh.Triangles, NewTriangle(
		[]mgl64.Vec3{{-halfW, h, -halfL}, {halfW, h, -halfL}, {halfW, 0, -halfL}},
	))
	// left
	triMesh.Triangles = append(triMesh.Triangles, NewTriangle(
		[]mgl64.Vec3{{-halfW, 0, -halfL}, {-halfW, 0, halfL}, {-halfW, h, halfL}},
	))
	triMesh.Triangles = append(triMesh.Triangles, NewTriangle(
		[]mgl64.Vec3{{-halfW, h, halfL}, {-halfW, h, -halfL}, {-halfW, 0, -halfL}},
	))
	// right
	triMesh.Triangles = append(triMesh.Triangles, NewTriangle(
		[]mgl64.Vec3{{halfW, 0, halfL}, {halfW, 0, -halfL}, {halfW, h, -halfL}},
	))
	triMesh.Triangles = append(triMesh.Triangles, NewTriangle(
		[]mgl64.Vec3{{halfW, h, -halfL}, {halfW, h, halfL}, {halfW, 0, halfL}},
	))
	// bottom
	triMesh.Triangles = append(triMesh.Triangles, NewTriangle(
		[]mgl64.Vec3{{halfW, 0, halfL}, {-halfW, 0, halfL}, {-halfW, 0, halfL}},
	))
	triMesh.Triangles = append(triMesh.Triangles, NewTriangle(
		[]mgl64.Vec3{{-halfW, 0, halfL}, {halfW, 0, -halfL}, {halfW, 0, halfL}},
	))
	// top
	triMesh.Triangles = append(triMesh.Triangles, NewTriangle(
		[]mgl64.Vec3{{-halfW, h, halfL}, {halfW, h, halfL}, {halfW, h, -halfL}},
	))
	triMesh.Triangles = append(triMesh.Triangles, NewTriangle(
		[]mgl64.Vec3{{halfW, h, -halfL}, {-halfW, h, -halfL}, {-halfW, h, halfL}},
	))
	return triMesh
}
