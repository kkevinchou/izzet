package navmesh

// func projectXZ(triangle Triangle, volume collider.BoundingBox, voxelDimension float64) {
// 	var verts [3]mgl64.Vec3
// 	verts[0], verts[1], verts[2] = triangle.ZSortedVerts[0], triangle.ZSortedVerts[1], triangle.ZSortedVerts[2]

// 	// drop y values
// 	verts[0][1] = 0
// 	verts[1][1] = 0
// 	verts[2][1] = 0

// 	// // left and right vertex
// 	// if verts[1][0] < verts[0][0] {
// 	// 	t := verts[1]
// 	// 	verts[1] = verts[0]
// 	// 	verts[0] = t
// 	// }

// 	// simplify with intercept theorem?
// 	// http://www.sunshine2k.de/coding/java/TriangleRasterization/TriangleRasterization.html
// 	midPoint := verts[1]
// 	edgeDir := verts[2].Sub(verts[0]).Normalize()
// 	horizontalMidpoint := verts[0].Add(edgeDir.Mul(midPoint[2] / edgeDir[2]))

// 	// ray0 := zSortedVerts[2].Sub(zSortedVerts[0]).Normalize()
// 	// ray0[1] = 0
// 	// ray1 := zSortedVerts[2].Sub(zSortedVerts[1]).Normalize()
// 	// ray1[1] = 0
// 	// var zFloat float64 = zSortedVerts[2].Z()

// 	// // yzField := [100][100]Span{}

// 	// vertRef0 := zSortedVerts[0]
// 	// vertRef0[1] = 0
// 	// vertRef1 := zSortedVerts[1]
// 	// vertRef1[1] = 0

// 	// for z := int(zSortedVerts[2].Z()); z >= int(zSortedVerts[0].Z()); z-- {
// 	// 	if int(zSortedVerts[1].Z()) == int(zFloat) {
// 	// 		ray1 = zSortedVerts[1].Sub(zSortedVerts[0]).Normalize()
// 	// 		ray1[1] = 0
// 	// 		vertRef1 = vertRef0
// 	// 	}

// 	// 	delta0 := zFloat - vertRef0.Z()
// 	// 	clippedVertex0 := vertRef0.Add(ray0.Mul(delta0 / ray0.Z()))

// 	// 	delta1 := zFloat - vertRef1.Z()
// 	// 	clippedVertex1 := vertRef1.Add(ray1.Mul(delta1 / ray1.Z()))

// 	// 	zFloat -= 1
// 	// }

// }

// // v1 - top
// // v2 - left
// // v3 - right

// // v2.y == v3.y
// func FillBottomFlatTriangle(v1, v2, v3 mgl64.Vec2) {
// 	invslope1 := mgl64.Vec2{v2.X() - v1.X(), v2.Y() - v1.Y()}.Normalize()
// 	invslope2 := mgl64.Vec2{v3.X() - v1.X(), v3.Y() - v1.Y()}.Normalize()

// 	// curx1 := v1.X()
// 	// curx2 := v1.X()

// 	upY1 := math.Ceil(v1.Y()) - v1.Y()
// 	adjustment1 := invslope1.Mul(upY1 / invslope1.Y())
// 	oneYUnitSlope1 := invslope1.Mul(1.0 / invslope1.Y())

// 	upY2 := math.Ceil(v2.Y()) - v2.Y()
// 	adjustment2 := invslope2.Mul(upY2 / invslope2.Y())
// 	oneYUnitSlope2 := invslope2.Mul(1.0 / invslope2.Y())

// 	// should be equal to V3 Y calculations since its a flat bottom triangle
// 	runCount := int(v1.Y()) - int(v2.Y())

// 	for i := 0; i < runCount; i++ {
// 		vert1 := v1.Add(invslope1.Mul(float64(i)))
// 		adjustedVertex1 := vert1.Add(adjustment1)
// 		nextAdjustedVertex1 := adjustedVertex1.Sub(oneYUnitSlope1)
// 		minX := math.Min(adjustedVertex1.X(), nextAdjustedVertex1.X())

// 		vert2 := v2.Add(invslope2.Mul(float64(i)))
// 		adjustedVertex2 := vert2.Add(adjustment2)
// 		nextAdjustedVertex2 := adjustedVertex2.Sub(oneYUnitSlope2)
// 		maxX := math.Max(adjustedVertex2.X(), nextAdjustedVertex2.X())
// 	}

// 	// for scanlineY := int(v1.Y()); scanlineY >= int(v2.Y()); scanlineY-- {
// 	// 	// drawLine(int(curx1), scanlineY, int(curx2), scanlineY)
// 	// 	v1.Y()
// 	// 	curx1 += invslope1
// 	// 	curx2 += invslope2
// 	// }
// }
