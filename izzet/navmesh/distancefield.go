package navmesh

const MaxDistanceFieldValue float64 = 999999999999999

func computeDistanceTransform(voxelField [][][]Voxel, reachField [][][]ReachInfo, dimensions [3]int) {
	// boundaries have a distance field of 0
	for y := 0; y < dimensions[1]; y++ {
		for x := 0; x < dimensions[0]; x++ {
			voxelField[x][y][0].DistanceField = 0
			voxelField[x][y][dimensions[2]-1].DistanceField = 0
		}
		for z := 0; z < dimensions[2]; z++ {
			voxelField[0][y][z].DistanceField = 0
			voxelField[dimensions[0]-1][y][z].DistanceField = 0
		}
	}

	for z := 1; z < dimensions[2]-1; z++ {
		for x := 1; x < dimensions[0]-1; x++ {
			for y := 0; y < dimensions[1]; y++ {
				computeVoxelDistanceTransform(x, y, z, voxelField, reachField, dimensions, -1)
			}
		}
	}

	for z := dimensions[2] - 2; z > 0; z-- {
		for x := dimensions[0] - 2; x > 0; x-- {
			for y := 0; y < dimensions[1]; y++ {
				computeVoxelDistanceTransform(x, y, z, voxelField, reachField, dimensions, 1)
			}
		}
	}
}

func computeVoxelDistanceTransform(x, y, z int, voxelField [][][]Voxel, reachField [][][]ReachInfo, dimensions [3]int, sweepDir int) {
	xNeighbor := voxelField[x+sweepDir][y][z]
	zNeighbor := voxelField[x][y][z+sweepDir]
	xzNegativeNeighbor := voxelField[x-1][y][z+sweepDir]
	xzPositiveNeighbor := voxelField[x+1][y][z+sweepDir]

	xReach := reachField[x+sweepDir][y][z]
	zReach := reachField[x][y][z+sweepDir]
	xzNegativeReach := reachField[x-1][y][z+sweepDir]
	xzPositiveReach := reachField[x+1][y][z+sweepDir]

	var minDistanceFieldValue float64 = MaxDistanceFieldValue

	if (!xNeighbor.Filled && !xReach.hasSource) || (!zNeighbor.Filled && !zReach.hasSource) || (!xzNegativeNeighbor.Filled && !xzNegativeReach.hasSource) || (!xzPositiveNeighbor.Filled && !xzPositiveReach.hasSource) {
		minDistanceFieldValue = 0
	} else {
		var distanceFieldXValue float64
		if xNeighbor.Filled {
			distanceFieldXValue = xNeighbor.DistanceField + 1
		} else if xReach.hasSource {
			source := xReach.source
			distanceFieldXValue = voxelField[source[0]][source[1]][source[2]].DistanceField + 1
		}

		var distanceFieldZValue float64
		if zNeighbor.Filled {
			distanceFieldZValue = zNeighbor.DistanceField + 1
		} else if zReach.hasSource {
			source := zReach.source
			distanceFieldZValue = voxelField[source[0]][source[1]][source[2]].DistanceField + 1
		}

		var distanceFieldXZNegativeValue float64
		if xzNegativeNeighbor.Filled {
			distanceFieldXZNegativeValue = xzNegativeNeighbor.DistanceField + 1.4
		} else if xzNegativeReach.hasSource {
			source := xzNegativeReach.source
			distanceFieldXZNegativeValue = voxelField[source[0]][source[1]][source[2]].DistanceField + 1.4
		}

		var distanceFieldXZPositiveValue float64
		if xzPositiveNeighbor.Filled {
			distanceFieldXZPositiveValue = xzPositiveNeighbor.DistanceField + 1.4
		} else if xzPositiveReach.hasSource {
			source := xzPositiveReach.source
			distanceFieldXZPositiveValue = voxelField[source[0]][source[1]][source[2]].DistanceField + 1.4
		}

		if distanceFieldXValue < minDistanceFieldValue {
			minDistanceFieldValue = distanceFieldXValue
		}
		if distanceFieldZValue < minDistanceFieldValue {
			minDistanceFieldValue = distanceFieldZValue
		}
		if distanceFieldXZNegativeValue < minDistanceFieldValue {
			minDistanceFieldValue = distanceFieldXZNegativeValue
		}
		if distanceFieldXZPositiveValue < minDistanceFieldValue {
			minDistanceFieldValue = distanceFieldXZPositiveValue
		}
	}

	if minDistanceFieldValue < voxelField[x][y][z].DistanceField {
		voxelField[x][y][z].DistanceField = minDistanceFieldValue
	}
}

var blurWeights []float64 = []float64{
	// 1, 2, 1,
	// 2, 4, 2,
	// 1, 2, 1,
	1, 1, 1,
	1, 1, 1,
	1, 1, 1,
}

func blurDistanceField(voxelField [][][]Voxel, reachField [][][]ReachInfo, dimensions [3]int) {
	blurredDistance := make([][][]float64, dimensions[0])
	for i := range voxelField {
		blurredDistance[i] = make([][]float64, dimensions[1])
		for j := range voxelField[i] {
			blurredDistance[i][j] = make([]float64, dimensions[2])
		}
	}

	for x := 0; x < dimensions[0]; x++ {
		for y := 0; y < dimensions[1]; y++ {
			for z := 0; z < dimensions[2]; z++ {

				var totalDistance float64
				var totalWeight float64
				for i, dir := range neighborDirs {
					if x+dir[0] < 0 || z+dir[1] < 0 || x+dir[0] >= dimensions[0] || z+dir[1] >= dimensions[2] {
						continue
					}

					var neighbor *Voxel
					reachNeighbor := &reachField[x+dir[0]][y][z+dir[1]]
					if reachNeighbor.sourceVoxel != nil {
						neighbor = reachNeighbor.sourceVoxel
					} else if voxelField[x+dir[0]][y][z+dir[1]].Filled {
						neighbor = &voxelField[x+dir[0]][y][z+dir[1]]
					}

					if neighbor == nil {
						continue
					}

					totalWeight += blurWeights[i]
					totalDistance += blurWeights[i] * neighbor.DistanceField
				}
				totalWeight += blurWeights[4]
				totalDistance += blurWeights[4] * voxelField[x][y][z].DistanceField

				totalDistance /= float64(totalWeight)
				blurredDistance[x][y][z] = totalDistance
			}
		}
	}

	for x := 0; x < dimensions[0]; x++ {
		for y := 0; y < dimensions[1]; y++ {
			for z := 0; z < dimensions[2]; z++ {
				voxelField[x][y][z].DistanceField = blurredDistance[x][y][z]
			}
		}
	}
}
