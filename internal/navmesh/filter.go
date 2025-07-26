package navmesh

import "math"

func ErodeWalkableArea(chf *CompactHeightField, erosionRadius float32) {
	distances := computeDistances(chf)

	minBoundaryDistance := int(erosionRadius * 2)
	for i := 0; i < chf.spanCount; i++ {
		if distances[i] < minBoundaryDistance {
			chf.areas[i] = NULL_AREA
		}
	}
}

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

				// the lowest difference between all spans that are reachable from the
				// current span. not considering slope
				lowestNeighborFloorDifference := maxHeight

				lowestTraversableNeighborFloor := span.Max
				highestTraversableNeighborFloor := span.Max

				for _, dir := range dirs {
					neighborX := x + xDirs[dir]
					neighborZ := z + zDirs[dir]

					// if neighbor spans are out of bounds, and mark the current span as a
					// ledge span
					if neighborX < 0 || neighborX >= xSize || neighborZ < 0 || neighborZ >= zSize {
						lowestNeighborFloorDifference = -climbableHeight - 1
						break
					}

					neighborSpan := hf.Spans[neighborX+neighborZ*xSize]
					neighborCeiling := maxHeight
					if neighborSpan != nil {
						neighborCeiling = neighborSpan.Min
					}

					// if there's a ceiling with a neighboring span, then there's a ledge the
					// agent can walk off of
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

						// if there isn't enough of a gap for the agent to walk through, this neighbor
						// cannot be considered a ledge, skip
						if min(ceiling, neighborCeiling)-max(floor, neighborFloor) < walkableHeight {
							continue
						}

						neighborFloorDifference := neighborFloor - floor
						// update the lowest neighbor floor distances, they may or may not be traversible
						lowestNeighborFloorDifference = min(lowestNeighborFloorDifference, neighborFloorDifference)

						// update the min/max traversable neighbor distances
						if math.Abs(float64(neighborFloorDifference)) <= float64(climbableHeight) {
							lowestTraversableNeighborFloor = min(lowestTraversableNeighborFloor, neighborFloor)
							highestTraversableNeighborFloor = max(highestTraversableNeighborFloor, neighborFloor)
						} else if neighborFloorDifference < -climbableHeight {
							// we already know this will be considered a ledge span so we can early-out
							break
						}
					}
				}

				if lowestNeighborFloorDifference < -climbableHeight {
					// the current span is close to a ledge if the magnitude of the drop to any neighbor span
					// is greater than the walkable climb distance
					span.Area = NULL_AREA
				} else if highestTraversableNeighborFloor-lowestTraversableNeighborFloor > climbableHeight {
					// if the difference between all neighbor floors is too high, we're on a steep slope
					// mark this current span as a ledge
					span.Area = NULL_AREA
				}
			}
		}
	}
}
