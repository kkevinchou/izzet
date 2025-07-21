package collider

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/kitolib/modelspec"
	"github.com/kkevinchou/kitolib/utils"
)

type Mesh interface {
	Vertices() []mgl64.Vec3
}

type Triangle struct {
	Normal mgl64.Vec3
	Points [3]mgl64.Vec3
}

func (t Triangle) Transform(transform mgl64.Mat4) Triangle {
	return NewTriangle([3]mgl64.Vec3{
		transform.Mul4x1(t.Points[0].Vec4(1)).Vec3(),
		transform.Mul4x1(t.Points[1].Vec4(1)).Vec3(),
		transform.Mul4x1(t.Points[2].Vec4(1)).Vec3(),
	})
}

func NewTriangle(points [3]mgl64.Vec3) Triangle {
	seg1 := points[1].Sub(points[0])
	seg2 := points[2].Sub(points[0])
	normal := seg1.Cross(seg2).Normalize()
	return Triangle{
		Points: points,
		Normal: normal,
	}
}

type TriMesh struct {
	Triangles   []Triangle
	DebugPoints []mgl64.Vec3
}

func (t TriMesh) Transform(transform mgl64.Mat4) TriMesh {
	newTriMesh := TriMesh{Triangles: make([]Triangle, len(t.Triangles))}
	for i, tri := range t.Triangles {
		newTriMesh.Triangles[i] = tri.Transform(transform)
	}
	return newTriMesh
}

// func NewTriMesh(model *model.Model) TriMesh {
// 	triMesh := TriMesh{}
// 	for _, mesh := range model.Meshes() {
// 		vertices := mesh.Vertices
// 		for i := 0; i < len(vertices); i += 3 {
// 			points := []mgl64.Vec3{
// 				utils.Vec3F32ToF64(vertices[i].Position),
// 				utils.Vec3F32ToF64(vertices[i+1].Position),
// 				utils.Vec3F32ToF64(vertices[i+2].Position),
// 			}
// 			triMesh.Triangles = append(triMesh.Triangles, NewTriangle(points))
// 		}
// 	}
// 	return triMesh
// }

func CreateTriMeshFromPrimitives(primitives []*modelspec.PrimitiveSpecification) *TriMesh {
	triMesh := TriMesh{}

	for _, p := range primitives {
		for i := 0; i < len(p.Vertices); i += 3 {
			v0 := p.Vertices[i]
			v1 := p.Vertices[i+1]
			v2 := p.Vertices[i+2]

			triVerts := [3]mgl64.Vec3{
				utils.Vec3F32ToF64(v0.Position),
				utils.Vec3F32ToF64(v1.Position),
				utils.Vec3F32ToF64(v2.Position),
			}

			vec1 := triVerts[1].Sub(triVerts[0])
			vec2 := triVerts[2].Sub(triVerts[1])

			tri := Triangle{
				Normal: vec1.Cross(vec2).Normalize(),
				Points: triVerts,
			}
			triMesh.Triangles = append(triMesh.Triangles, tri)
		}
	}

	return &triMesh
}
