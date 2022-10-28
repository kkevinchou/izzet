package libutils

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
)

func MatF32FromColumnMajorFloats(matrix [16]float32) mgl32.Mat4 {
	return mgl32.Mat4FromRows(
		mgl32.Vec4{matrix[0], matrix[1], matrix[2], matrix[3]},
		mgl32.Vec4{matrix[4], matrix[5], matrix[6], matrix[7]},
		mgl32.Vec4{matrix[8], matrix[9], matrix[10], matrix[11]},
		mgl32.Vec4{matrix[12], matrix[13], matrix[14], matrix[15]},
	)
}

func Vec3ApproxEqualZero(v mgl64.Vec3) bool {
	return Vec3ApproxEqualThreshold(v, mgl64.Vec3{}, 1)
}

func Vec3ApproxEqualThreshold(v1 mgl64.Vec3, v2 mgl64.Vec3, threshold float64) bool {
	return v1.ApproxFuncEqual(v2, func(a, b float64) bool {
		return math.Abs(a-b) < threshold
	})
}

func Vec3IsZero(v mgl64.Vec3) bool {
	return v[0] == 0 && v[1] == 0 && v[2] == 0
}

func Cross2D(v1, v2 mgl64.Vec3) float64 {
	return (v1.X() * v2.Z()) - (v1.Z() * v2.X())
}

func Vec3F32ToF64(v mgl32.Vec3) mgl64.Vec3 {
	var result mgl64.Vec3
	for i := 0; i < len(v); i++ {
		result[i] = float64(v[i])
	}
	return result
}

func Vec3F64ToF32(v mgl64.Vec3) mgl32.Vec3 {
	var result mgl32.Vec3
	for i := 0; i < len(v); i++ {
		result[i] = float32(v[i])
	}
	return result
}

func Vec4F64To4F32(v mgl64.Vec4) mgl32.Vec4 {
	return mgl32.Vec4{float32(v.X()), float32(v.Y()), float32(v.Z()), float32(v.W())}
}

func QuatF64ToF32(q mgl64.Quat) mgl32.Quat {
	return mgl32.Quat{W: float32(q.W), V: Vec3F64ToF32(q.V)}
}

func Mat4F64ToF32(m mgl64.Mat4) mgl32.Mat4 {
	var result mgl32.Mat4
	for i := 0; i < len(m); i++ {
		result[i] = float32(m[i])
	}
	return result
}

func NormalizeF64(a float64) float64 {
	if a > 0 {
		return 1
	} else if a < 0 {
		return -1
	}
	return 0
}

func SameSign(a, b float64) bool {
	return (a > 0 && b > 0) || (a < 0 && b < 0) || (a == 0 && b == 0)
}

// takes a matrix composed of a translation, scale, and rotation (no projection) and decomposes it
func Decompose(m mgl32.Mat4) (mgl32.Vec3, mgl32.Quat, mgl32.Vec3) {
	translation := m.Col(3).Vec3()
	m.SetCol(3, mgl32.Vec4{0, 0, 0, 1})

	xScaleCol := m.Col(0).Vec3()
	newXCol := xScaleCol.Mul(1. / xScaleCol.Len())
	yScaleCol := m.Col(1).Vec3()
	newYCol := yScaleCol.Mul(1. / yScaleCol.Len())
	zScaleCol := m.Col(2).Vec3()
	newZCol := zScaleCol.Mul(1. / zScaleCol.Len())
	m.SetCol(0, newXCol.Vec4(0))
	m.SetCol(1, newYCol.Vec4(0))
	m.SetCol(2, newZCol.Vec4(0))

	rotation := mgl32.Mat4ToQuat(m)
	scale := mgl32.Vec3{xScaleCol.Len(), yScaleCol.Len(), zScaleCol.Len()}

	return translation, rotation, scale
}

// Quaternion interpolation, reimplemented from: https://github.com/TheThinMatrix/OpenGL-Animation/blob/dde792fe29767192bcb60d30ac3e82d6bcff1110/Animation/animation/Quaternion.java#L158
func QInterpolate(a, b mgl32.Quat, blend float32) mgl32.Quat {
	var result mgl32.Quat = mgl32.Quat{}
	var dot float32 = a.W*b.W + a.V.X()*b.V.X() + a.V.Y()*b.V.Y() + a.V.Z()*b.V.Z()
	blendI := float32(1) - blend
	if dot < 0 {
		result.W = blendI*a.W + blend*-b.W
		result.V = mgl32.Vec3{
			blendI*a.V.X() + blend*-b.V.X(),
			blendI*a.V.Y() + blend*-b.V.Y(),
			blendI*a.V.Z() + blend*-b.V.Z(),
		}
	} else {
		result.W = blendI*a.W + blend*b.W
		result.V = mgl32.Vec3{
			blendI*a.V.X() + blend*b.V.X(),
			blendI*a.V.Y() + blend*b.V.Y(),
			blendI*a.V.Z() + blend*b.V.Z(),
		}
	}

	return result.Normalize()
}

// Quaternion interpolation, reimplemented from: https://github.com/TheThinMatrix/OpenGL-Animation/blob/dde792fe29767192bcb60d30ac3e82d6bcff1110/Animation/animation/Quaternion.java#L158
func QInterpolate64(a, b mgl64.Quat, blend float64) mgl64.Quat {
	var result mgl64.Quat = mgl64.Quat{}
	var dot float64 = a.W*b.W + a.V.X()*b.V.X() + a.V.Y()*b.V.Y() + a.V.Z()*b.V.Z()
	blendI := float64(1) - blend
	if dot < 0 {
		result.W = blendI*a.W + blend*-b.W
		result.V = mgl64.Vec3{
			blendI*a.V.X() + blend*-b.V.X(),
			blendI*a.V.Y() + blend*-b.V.Y(),
			blendI*a.V.Z() + blend*-b.V.Z(),
		}
	} else {
		result.W = blendI*a.W + blend*b.W
		result.V = mgl64.Vec3{
			blendI*a.V.X() + blend*b.V.X(),
			blendI*a.V.Y() + blend*b.V.Y(),
			blendI*a.V.Z() + blend*b.V.Z(),
		}
	}

	return result.Normalize()
}

func Vec3ToQuat(v mgl64.Vec3) mgl64.Quat {
	return mgl64.QuatBetweenVectors(mgl64.Vec3{0, 0, -1}, v)
}
