package navmesh

const (
	blurSkipDistance = 2
)

func BuildDistanceField(chf *CompactHeightField) {
	width := chf.width
	height := chf.height

	distances := make([]int, chf.spanCount)
	for i := range distances {
		distances[i] = maxDistance
	}

	// mark boundary spans
	for z := range height {
		for x := range width {
			cell := chf.cells[x+z*width]
			spanIndex := cell.SpanIndex
			spanCount := cell.SpanCount

			for i := spanIndex; i < spanIndex+SpanIndex(spanCount); i++ {
				span := chf.spans[i]
				for _, neighborIndex := range span.neighbors {
					if neighborIndex == -1 {
						distances[i] = 0
						break
					}
				}
			}
		}
	}

	// pass 1
	for z := range height {
		for x := range width {
			cell := chf.cells[x+z*width]
			spanIndex := cell.SpanIndex
			spanCount := cell.SpanCount

			for i := spanIndex; i < spanIndex+SpanIndex(spanCount); i++ {
				span := chf.spans[i]

				// (-1, 0)
				if span.neighbors[dirLeft] != -1 {
					neighborSpan := chf.spans[span.neighbors[dirLeft]]
					distances[i] = Min(distances[i], distances[span.neighbors[dirLeft]]+2)

					// (-1, -1)
					if neighborSpan.neighbors[dirUp] != -1 {
						distances[i] = Min(distances[i], distances[neighborSpan.neighbors[dirUp]]+3)
					}
				}

				// (0, -1)
				if span.neighbors[dirUp] != -1 {
					neighborSpan := chf.spans[span.neighbors[dirUp]]
					distances[i] = Min(distances[i], distances[span.neighbors[dirUp]]+2)

					// (1, -1)
					if neighborSpan.neighbors[dirRight] != -1 {
						distances[i] = Min(distances[i], distances[neighborSpan.neighbors[dirRight]]+3)
					}
				}
			}
		}
	}

	// pass 2
	for z := height - 1; z >= 0; z-- {
		for x := height - 1; x >= 0; x-- {
			cell := chf.cells[x+z*width]
			spanIndex := cell.SpanIndex
			spanCount := cell.SpanCount

			for i := spanIndex; i < spanIndex+SpanIndex(spanCount); i++ {
				span := chf.spans[i]

				// (1, 0)
				if span.neighbors[dirRight] != -1 {
					neighborSpan := chf.spans[span.neighbors[dirRight]]
					distances[i] = Min(distances[i], distances[span.neighbors[dirRight]]+2)

					// (1, 1)
					if neighborSpan.neighbors[dirDown] != -1 {
						distances[i] = Min(distances[i], distances[neighborSpan.neighbors[dirDown]]+3)
					}
				}

				// (0, 1)
				if span.neighbors[dirDown] != -1 {
					neighborSpan := chf.spans[span.neighbors[dirDown]]
					distances[i] = Min(distances[i], distances[span.neighbors[dirDown]]+2)

					// (-1, 1)
					if neighborSpan.neighbors[dirLeft] != -1 {
						distances[i] = Min(distances[i], distances[neighborSpan.neighbors[dirLeft]]+3)
					}
				}
			}
		}
	}

	maxDist := 0
	for i := 0; i < chf.spanCount; i++ {
		maxDist = Max(distances[i], maxDist)
	}

	chf.Distances = BoxBlur(chf, distances)
	chf.maxDistance = maxDist
}

func BoxBlur(chf *CompactHeightField, distances []int) []int {
	width := chf.width
	height := chf.height

	blurredDistances := make([]int, chf.spanCount)

	for z := range height {
		for x := range width {
			cell := chf.cells[x+z*width]
			spanIndex := cell.SpanIndex
			spanCount := cell.SpanCount

			for i := spanIndex; i < spanIndex+SpanIndex(spanCount); i++ {
				cellDistance := distances[i]
				span := chf.spans[i]

				if cellDistance <= blurSkipDistance {
					// if we're close to a wall, skip blur
					blurredDistances[i] = cellDistance
					continue
				}

				totalDistance := cellDistance
				for dir := range dirs {
					neighborSpanIndex := span.neighbors[dir]
					if neighborSpanIndex != -1 {
						neighborSpan := chf.spans[neighborSpanIndex]
						totalDistance += distances[neighborSpanIndex]

						diagNeighborDir := (dir + 1) % 4
						diagNeighborSpanIndex := neighborSpan.neighbors[diagNeighborDir]
						if diagNeighborSpanIndex != -1 {
							totalDistance += distances[diagNeighborSpanIndex]
						} else {
							totalDistance += cellDistance
						}
					} else {
						totalDistance += 2 * cellDistance
					}
				}

				blurredDistances[i] = (totalDistance + 5) / 9
			}
		}
	}
	return blurredDistances
}
