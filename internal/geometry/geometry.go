package geometry

import (
	"github.com/go-gl/mathgl/mgl64"
)

type Plane struct {
	A, B, C, D float64
}

func ComputeErrorQuadric(plane Plane) mgl64.Mat4 {
	a := plane.A
	b := plane.B
	c := plane.C
	d := plane.D

	// quadric := mgl64.Mat4FromCols(
	// 	mgl64.Vec4{a * a, a * b, a * c, a * d},
	// 	mgl64.Vec4{a * b, b * b, b * c, b * d},
	// 	mgl64.Vec4{a * c, b * c, c * c, c * d},
	// 	mgl64.Vec4{a * d, b * d, c * d, d * d},
	// )

	v := mgl64.Vec4{a, b, c, d}
	quadric := v.OuterProd4(v)

	return quadric
}

func PlaneFromVerts(verts [3]mgl64.Vec3) (Plane, bool) {
	v1 := verts[1].Sub(verts[0])
	v2 := verts[2].Sub(verts[1])

	normal := v1.Cross(v2)
	if normal.Len() == 0 {
		return Plane{}, false
	}
	normal = normal.Normalize()

	samplePoint := verts[0]
	d := -(samplePoint[0]*normal[0] + samplePoint[1]*normal[1] + samplePoint[2]*normal[2])
	return Plane{A: normal[0], B: normal[1], C: normal[2], D: d}, true
}

func ComputeQEM(v mgl64.Vec4, q mgl64.Mat4) float64 {
	return v.Dot(q.Mul4x1(v))
}

// func PreMultiply(v mgl64.Vec4, m mgl64.Mat4) mgl64.Vec4 {
// 	return mgl64.Vec4{
// 		v[0] * (m[0] + m[1] + m[2] + m[3]),
// 		v[1] * (m[4] + m[5] + m[6] + m[7]),
// 		v[2] * (m[8] + m[9] + m[10] + m[11]),
// 		v[3] * (m[12] + m[13] + m[14] + m[15]),
// 	}
// }
