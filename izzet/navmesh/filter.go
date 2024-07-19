package navmesh

import "math"

func FilterLowHeightSpans(walkableHeight int, hf *HeightField) {
	xSize := hf.Width
	zSize := hf.Height
	for z := range zSize {
		for x := range xSize {
			for span := hf.Spans[x+z*xSize]; span != nil; span = span.Next {
				floor := span.Max
				ceiling := maxHeight
				if span.Next != nil {
					ceiling = span.Next.Min
				}
				if ceiling-floor < walkableHeight {
					span.Area = NULL_AREA
				}
			}
		}
	}
}

func FilterLedgeSpans(walkableHeight, climbableHeight int, hf *HeightField) {
	xSize := hf.Width
	zSize := hf.Height

	for z := range zSize {
		for x := range xSize {
			for span := hf.Spans[x+z*xSize]; span != nil; span = span.Next {
				if span.Area == NULL_AREA {
					continue
				}

				floor := span.Max
				ceiling := maxHeight
				if span.Next != nil {
					ceiling = span.Next.Min
				}

				lowestNeighborFloorDifference := maxHeight

				lowestTraversableNeighborFloor := span.Max
				highestTraversableNeighborFloor := span.Max

				for _, dir := range dirs {
					neighborX := x + xDirs[dir]
					neighborZ := z + zDirs[dir]

					if neighborX < 0 || neighborX >= xSize || neighborZ < 0 || neighborZ >= zSize {
						lowestNeighborFloorDifference = -climbableHeight - 1
						break
					}

					neighborSpan := hf.Spans[neighborX+neighborZ*xSize]
					neighborCeiling := maxHeight
					if neighborSpan != nil {
						neighborCeiling = neighborSpan.Min
					}

					if min(ceiling, neighborCeiling)-floor >= walkableHeight {
						lowestNeighborFloorDifference = (-climbableHeight) - 1
						break
					}

					for ; neighborSpan != nil; neighborSpan = neighborSpan.Next {
						neighborFloor := neighborSpan.Max
						neighborCeiling := maxHeight
						if neighborSpan.Next != nil {
							neighborCeiling = neighborSpan.Next.Min
						}

						if min(ceiling, neighborCeiling)-max(floor, neighborFloor) < climbableHeight {
							continue
						}

						neighborFloorDifference := neighborFloor - floor
						lowestNeighborFloorDifference = min(lowestNeighborFloorDifference, neighborFloorDifference)

						if math.Abs(float64(neighborFloorDifference)) <= float64(climbableHeight) {
							lowestNeighborFloorDifference = min(lowestTraversableNeighborFloor, neighborFloor)
							highestTraversableNeighborFloor = max(highestTraversableNeighborFloor, neighborFloor)
						} else if neighborFloorDifference < -climbableHeight {
							break
						}
					}
				}

				if lowestNeighborFloorDifference < -climbableHeight {
					span.Area = NULL_AREA
				} else if highestTraversableNeighborFloor-lowestTraversableNeighborFloor > climbableHeight {
					span.Area = NULL_AREA
				}
			}
		}
	}
}
