package navmesh

import "fmt"

func BuildRegions(chf *CompactHeightField, iterationCount int) {
	const maxStacks int = 8

	level := (chf.maxDistance + 1) & ^1
	stackID := -1
	stacks := make([][]LevelStackEntry, maxStacks)

	regions := make([]int, chf.spanCount)
	distances := make([]int, chf.spanCount)
	regionID := 1

	currentIterationCount := 0

	for level > 0 {
		level -= 2
		if level < 0 {
			level = 0
		}
		stackID = (stackID + 1) % maxStacks

		if stackID == 0 {
			sortCellsByLevel(level, chf, regions, maxStacks, stacks)
			var minDist int = 9999999
			var maxDist int = -999999
			for i := 0; i < len(stacks[0]); i++ {
				minDist = min(minDist, stacks[0][i].distance)
				maxDist = max(maxDist, stacks[0][i].distance)
			}
			fmt.Println("NEW SORT MIN", minDist, "MAX", maxDist)
		} else {
			appendStacks(stacks[stackID-1], &stacks[stackID], regions)
		}

		expandRegions(level, chf, stacks[stackID], distances, regions, false)

		currentIterationCount++
		if currentIterationCount >= iterationCount {
			break
		}

		for _, entry := range stacks[stackID] {
			if entry.spanIndex >= 0 && regions[entry.spanIndex] == 0 {
				if floodRegion(entry.x, entry.z, entry.spanIndex, level, regionID, chf, regions, distances) {
					if regionID == 0xFFFFF {
						panic("HIT MAX REGIONS")
					}
					regionID++
				}
			}
		}

		currentIterationCount++
		if currentIterationCount >= iterationCount {
			break
		}

		count := 0
		total := 0
		for i := 0; i < len(stacks[stackID]); i++ {
			idx := stacks[stackID][i].spanIndex
			if idx == -1 {
				continue
			}

			total++
			if regions[idx] != 0 {
				count++
			}
		}
		fmt.Printf("\t %d/%d entries assigned a region\n", count, len(stacks[stackID]))
	}

	count := 0
	total := 0
	for i := 0; i < len(stacks[stackID]); i++ {
		idx := stacks[stackID][i].spanIndex
		if idx == -1 {
			continue
		}

		total++
		if regions[idx] != 0 {
			count++
		}
	}
	fmt.Printf("\t %d/%d entries assigned a region\n", count, len(stacks[stackID]))

	// expandRegions(64, 0, chf, nil, distances, regions, true)

	// merge and filter regions

	for i := range chf.spanCount {
		chf.spans[i].regionID = regions[i]
	}
}

// floodRegion creates new seed regions (distance 0)
// find all the neighboring spans of the current level and set their region id
// if a span has a neighbor that has a different region id than itself, unset its span id
func floodRegion(x, z int, spanIndex SpanIndex, level, regionID int, chf *CompactHeightField, regions, distances []int) bool {
	lev := level - 2
	if lev < 0 {
		lev = 0
	}

	var queue []LevelStackEntry
	queue = append(queue, LevelStackEntry{x: x, z: z, spanIndex: spanIndex})
	regions[spanIndex] = regionID
	distances[spanIndex] = 0

	var count int
	area := chf.areas[spanIndex]

	for len(queue) > 0 {
		entry := queue[0]
		queue = queue[1:]

		spanIndex := entry.spanIndex
		span := chf.spans[spanIndex]
		differingNeighborRegion := false

		for _, dir := range dirs {
			neighborSpanIndex := span.neighbors[dir]
			if neighborSpanIndex == -1 {
				continue
			}

			if chf.areas[neighborSpanIndex] != area {
				continue
			}

			if regions[neighborSpanIndex] != 0 && regions[neighborSpanIndex] != regions[spanIndex] {
				differingNeighborRegion = true
				break
			}

			dir2 := (dir + 1) % 4
			neighborSpan := chf.spans[neighborSpanIndex]
			neighborSpanIndex2 := neighborSpan.neighbors[dir2]
			if neighborSpanIndex2 == -1 {
				continue
			}

			if chf.areas[neighborSpanIndex2] != area {
				continue
			}

			if regions[neighborSpanIndex2] != 0 && regions[neighborSpanIndex2] != regions[spanIndex] {
				differingNeighborRegion = true
				break
			}
		}

		if differingNeighborRegion {
			regions[spanIndex] = 0
			continue
		}

		count++

		for _, dir := range dirs {
			neighborSpanIndex := span.neighbors[dir]
			if neighborSpanIndex == -1 {
				continue
			}

			if chf.areas[neighborSpanIndex] != area {
				continue
			}

			if chf.distances[neighborSpanIndex] >= lev && regions[neighborSpanIndex] == 0 {
				regions[neighborSpanIndex] = regionID
				distances[neighborSpanIndex] = 0
				queue = append(queue, LevelStackEntry{x: x + xDirs[dir], z: z + zDirs[dir], spanIndex: neighborSpanIndex, distance: chf.distances[neighborSpanIndex]})
			}
		}
	}

	return count > 0
}

const levelCoalescing int = 4

func sortCellsByLevel(startLevel int, chf *CompactHeightField, regions []int, maxStacks int, stacks [][]LevelStackEntry) {
	startLevel = startLevel / levelCoalescing

	for i := range maxStacks {
		stacks[i] = nil
	}

	for z := range chf.height {
		for x := range chf.width {
			cell := &chf.cells[x+z*chf.width]
			spanIndex := cell.SpanIndex
			spanCount := cell.SpanCount

			for i := spanIndex; i < spanIndex+SpanIndex(spanCount); i++ {
				// skip spans that have a region assigned already
				if regions[i] != 0 && chf.areas[i] == 0 {
					continue
				}

				level := chf.distances[i] / levelCoalescing
				stackID := startLevel - level

				if stackID >= maxStacks {
					continue
				}

				if stackID < 0 {
					stackID = 0
				}

				stacks[stackID] = append(stacks[stackID], LevelStackEntry{x: x, z: z, spanIndex: i, distance: chf.distances[i]})
			}
		}
	}
}

func appendStacks(src []LevelStackEntry, dest *[]LevelStackEntry, regions []int) {
	for _, entry := range src {
		if entry.spanIndex < 0 || regions[entry.spanIndex] != 0 {
			continue
		}
		*dest = append(*dest, entry)
	}
}

// expandRegions iterates through each entry in the current stack and assigns a region
// to it based on the neighbor that has the smallest distance from a region seed
func expandRegions(level int, chf *CompactHeightField, stack []LevelStackEntry, distances []int, regions []int, fill bool) {
	totalSpans := 0
	if fill {
		stack = nil
		for z := range chf.height {
			for x := range chf.width {
				cell := &chf.cells[x+z*chf.width]
				spanIndex := cell.SpanIndex
				spanCount := cell.SpanCount
				totalSpans += spanCount
				for i := spanIndex; i < spanIndex+SpanIndex(spanCount); i++ {
					if chf.distances[i] >= level && regions[i] == 0 && chf.areas[i] != NULL_AREA {
						stack = append(stack, LevelStackEntry{x: x, z: z, spanIndex: i})
					}
				}
			}
		}
	} else {
		for i := range stack {
			if regions[stack[i].spanIndex] != 0 {
				stack[i].spanIndex = -1
			}
		}
	}

	failed := 0
	var dirtyEntries []DirtyEntry

	for len(stack) > 0 {
		dirtyEntries = nil
		failed = 0

		// repeatedly go through the entire stack and attempt to assign regions to each span
		// if all spans fail to assign a region, we're done.
		// NOTE: does the order in which spans are assigned regions matter?
		// 	seems like if we assign a region early on, we don't have the opportunity to assign
		// 	a closer region at a later iteration
		for j := 0; j < len(stack); j++ {
			spanIndex := stack[j].spanIndex
			if spanIndex < 0 {
				failed++
				continue
			}

			newDistance := maxDistance
			span := chf.spans[spanIndex]
			regionID := regions[spanIndex]
			area := chf.areas[spanIndex]

			for _, dir := range dirs {
				neighborSpanIndex := span.neighbors[dir]
				if neighborSpanIndex == -1 {
					continue
				}

				if chf.areas[neighborSpanIndex] != area {
					continue
				}

				if regions[neighborSpanIndex] > 0 {
					if distances[neighborSpanIndex]+2 < newDistance {
						regionID = regions[neighborSpanIndex]
						newDistance = distances[neighborSpanIndex] + 2
					}
				}
			}

			if regionID > 0 {
				stack[j].spanIndex = -1
				dirtyEntries = append(dirtyEntries, DirtyEntry{spanIndex: spanIndex, regionID: regionID, distance: newDistance})
			} else {
				failed++
			}
		}

		for _, entry := range dirtyEntries {
			idx := entry.spanIndex
			regions[idx] = entry.regionID
			distances[idx] = entry.distance
		}

		if failed == len(stack) {
			break
		}
	}
}

type SpanIndex int

type LevelStackEntry struct {
	x, z      int
	spanIndex SpanIndex
	distance  int
}

type DirtyEntry struct {
	spanIndex SpanIndex
	regionID  int
	distance  int
}
