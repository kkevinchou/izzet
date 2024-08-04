package render

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/kkevinchou/izzet/izzet/settings"
)

func computeCubeMapTransforms(position mgl32.Vec3, near, far float32) []mgl32.Mat4 {
	projectionMatrix := mgl32.Perspective(mgl32.DegToRad(90), float32(settings.DepthCubeMapWidth)/float32(settings.DepthCubeMapHeight), near, far)

	cubeMapTransforms := []mgl32.Mat4{
		projectionMatrix.Mul4( // right
			mgl32.LookAtV(
				position,
				position.Add(mgl32.Vec3{1, 0, 0}),
				mgl32.Vec3{0, -1, 0},
			),
		),
		projectionMatrix.Mul4( // left
			mgl32.LookAtV(
				position,
				position.Add(mgl32.Vec3{-1, 0, 0}),
				mgl32.Vec3{0, -1, 0},
			),
		),
		projectionMatrix.Mul4( // up
			mgl32.LookAtV(
				position,
				position.Add(mgl32.Vec3{0, 1, 0}),
				mgl32.Vec3{0, 0, 1},
			),
		),
		projectionMatrix.Mul4( // down
			mgl32.LookAtV(
				position,
				position.Add(mgl32.Vec3{0, -1, 0}),
				mgl32.Vec3{0, 0, -1},
			),
		),
		projectionMatrix.Mul4( // back
			mgl32.LookAtV(
				position,
				position.Add(mgl32.Vec3{0, 0, 1}),
				mgl32.Vec3{0, -1, 0},
			),
		),
		projectionMatrix.Mul4( // front
			mgl32.LookAtV(
				position,
				position.Add(mgl32.Vec3{0, 0, -1}),
				mgl32.Vec3{0, -1, 0},
			),
		),
	}
	return cubeMapTransforms
}
