package render

// func computeCubeMapTransforms(position mgl64.Vec3, near, far float64) []mgl64.Mat4 {
// 	projectionMatrix := mgl64.Perspective(mgl64.DegToRad(90), float64(settings.DepthCubeMapWidth)/float64(settings.DepthCubeMapHeight), near, far)

// 	cubeMapTransforms := []mgl64.Mat4{
// 		projectionMatrix.Mul4( // right
// 			mgl64.LookAtV(
// 				position,
// 				position.Add(mgl64.Vec3{1, 0, 0}),
// 				mgl64.Vec3{0, -1, 0},
// 			),
// 		),
// 		projectionMatrix.Mul4( // left
// 			mgl64.LookAtV(
// 				position,
// 				position.Add(mgl64.Vec3{-1, 0, 0}),
// 				mgl64.Vec3{0, -1, 0},
// 			),
// 		),
// 		projectionMatrix.Mul4( // up
// 			mgl64.LookAtV(
// 				position,
// 				position.Add(mgl64.Vec3{0, 1, 0}),
// 				mgl64.Vec3{0, 0, 1},
// 			),
// 		),
// 		projectionMatrix.Mul4( // down
// 			mgl64.LookAtV(
// 				position,
// 				position.Add(mgl64.Vec3{0, -1, 0}),
// 				mgl64.Vec3{0, 0, -1},
// 			),
// 		),
// 		projectionMatrix.Mul4( // back
// 			mgl64.LookAtV(
// 				position,
// 				position.Add(mgl64.Vec3{0, 0, 1}),
// 				mgl64.Vec3{0, -1, 0},
// 			),
// 		),
// 		projectionMatrix.Mul4( // front
// 			mgl64.LookAtV(
// 				position,
// 				position.Add(mgl64.Vec3{0, 0, -1}),
// 				mgl64.Vec3{0, -1, 0},
// 			),
// 		),
// 	}
// 	return cubeMapTransforms
// }
