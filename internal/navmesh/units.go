package navmesh

import "math"

const voxelConversionEpsilon = 1e-5

func WorldRadiusToVoxels(radius, cellSize float32) int {
	if radius <= 0 || cellSize <= 0 {
		return 0
	}
	return int(math.Ceil(float64(radius/cellSize) - voxelConversionEpsilon))
}

func WorldHeightToVoxels(height, cellHeight float32) int {
	if height <= 0 || cellHeight <= 0 {
		return 0
	}
	return int(math.Ceil(float64(height/cellHeight) - voxelConversionEpsilon))
}

func WorldClimbToVoxels(height, cellHeight float32) int {
	if height <= 0 || cellHeight <= 0 {
		return 0
	}
	return int(math.Floor(float64(height/cellHeight) + voxelConversionEpsilon))
}
