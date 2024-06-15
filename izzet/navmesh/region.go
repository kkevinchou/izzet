package navmesh

const (
	levelStepSize = 2
)

func BuildRegions(chf *CompactHeightField, iterationCount int, minRegionArea int, mergeRegionSize int) {
	const maxStacks int = 8

	// round the level up to the nearest even number
	level := (chf.maxDistance + 1) & ^1
	stackID := -1
	stacks := make([][]LevelStackEntry, maxStacks)

	regionIDs := make([]int, chf.spanCount)
	distances := make([]int, chf.spanCount)
	regionID := 1

	currentIterationCount := 0

	for level > 0 {
		level -= levelStepSize
		if level < 0 {
			level = 0
		}
		stackID = (stackID + 1) % maxStacks

		if stackID == 0 {
			sortCellsByLevel(level, chf, regionIDs, maxStacks, stacks)
		} else {
			appendStacks(stacks[stackID-1], &stacks[stackID], regionIDs)
		}

		expandRegions(level, chf, stacks[stackID], distances, regionIDs, false)

		currentIterationCount++
		if currentIterationCount >= iterationCount {
			break
		}

		for _, entry := range stacks[stackID] {
			if entry.spanIndex >= 0 && regionIDs[entry.spanIndex] == 0 {
				if floodRegion(entry.x, entry.z, entry.spanIndex, level, regionID, chf, regionIDs, distances) {
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
			if regionIDs[idx] != 0 {
				count++
			}
		}
	}
	// we pre-emptively increment regionID so when we're done we should decrement it
	chf.maxRegions = regionID - 1

	// expandRegions(0, chf, nil, distances, regions, true)

	mergeAndFilterRegions(chf, regionIDs, minRegionArea, mergeRegionSize, &chf.maxRegions)

	// update the region id for each span
	for i := range chf.spanCount {
		chf.spans[i].regionID = regionIDs[i]
	}

}

// floodRegion creates new seed regions (distance 0)
// find all the neighboring spans of the current level and set their region id
// if a span has a neighbor that has a different region id than itself, unset its span id
func floodRegion(x, z int, spanIndex SpanIndex, level, regionID int, chf *CompactHeightField, regions, distances []int) bool {
	lev := level - levelStepSize
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

func sortCellsByLevel(startLevel int, chf *CompactHeightField, regions []int, maxStacks int, stacks [][]LevelStackEntry) {
	startLevel = startLevel / levelStepSize

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
				if regions[i] != 0 || chf.areas[i] == NULL_AREA {
					continue
				}

				level := chf.distances[i] / levelStepSize
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

// mergeAndFilterRegions mutates regionIDs and maxRegionID
func mergeAndFilterRegions(chf *CompactHeightField, regionIDs []int, minRegionArea int, mergeRegionSize int, maxRegionID *int) int {
	numRegions := *maxRegionID + 1
	regions := make([]Region, numRegions)
	for i := 0; i < numRegions; i++ {
		regions[i].id = i
	}

	// find the edge of a region and find connections around the contour
	for z := range chf.height {
		for x := range chf.width {
			cell := &chf.cells[x+z*chf.width]
			spanIndex := cell.SpanIndex
			spanCount := cell.SpanCount
			for i := spanIndex; i < spanIndex+SpanIndex(spanCount); i++ {
				regionID := regionIDs[i]
				region := &regions[regionID]
				if regionID == 0 || regionID >= numRegions {
					continue
				}

				region.spanCount++

				for j := spanIndex; j < spanIndex+SpanIndex(spanCount); j++ {
					if i == j {
						continue
					}
					floorID := regionIDs[j]
					if floorID == 0 || floorID >= numRegions {
						continue
					}
					if floorID == regionID {
						region.overlap = true
					}
					addUniqueFloorRegion(region, floorID)
				}

				region.areaType = chf.areas[i]

				edgeDir := -1
				for _, dir := range dirs {
					if isSolidEdge(chf, regionIDs, spanIndex, dir) {
						edgeDir = dir
						break
					}
				}

				if edgeDir != -1 {
					region.connections = walkContour(chf, spanIndex, edgeDir, regionIDs)
				}
			}
		}
	}

	// remove regions that are too small

	var stack []int
	var trace []int

	for _, region := range regions {
		if region.id == 0 {
			continue
		}
		if region.spanCount == 0 {
			continue
		}
		if region.visited {
			continue
		}

		stack = []int{region.id}
		trace = nil
		spanCount := 0
		region.visited = true

		for len(stack) > 0 {
			currentRegionID := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			currentRegion := &regions[currentRegionID]

			spanCount += currentRegion.spanCount
			trace = append(trace, currentRegionID)

			for _, conn := range currentRegion.connections {
				neighborRegion := &regions[conn]
				if neighborRegion.visited {
					continue
				}
				if neighborRegion.id == 0 {
					continue
				}
				stack = append(stack, neighborRegion.id)
				neighborRegion.visited = true
			}
		}

		// remove clusters of regions that are too small
		if spanCount < minRegionArea {
			for _, regionID := range trace {
				regions[regionID].spanCount = 0
				regions[regionID].id = 0
			}
		}
	}

	// merge small regions to neighbor regions

	// compress region IDs

	// mark valid regions we want to remap to the front
	for i := range numRegions {
		region := &regions[i]
		region.remap = false
		if region.id == 0 {
			continue
		}
		region.remap = true
	}

	regionIDGen := 0
	for i := range numRegions {
		region := &regions[i]
		if !region.remap {
			continue
		}

		oldID := region.id
		regionIDGen++
		newID := regionIDGen

		for j := i; j < numRegions; j++ {
			curRegion := &regions[j]
			if curRegion.id == oldID {
				curRegion.id = newID
				curRegion.remap = false
			}
		}
	}
	*maxRegionID = regionIDGen
	for i := range chf.spanCount {
		regionIDs[i] = regions[regionIDs[i]].id
	}

	// remap regions

	// return overlapping regions

	return 0
}

func walkContour(chf *CompactHeightField, spanIndex SpanIndex, dir int, regionIDs []int) []int {
	startDir := dir
	startIndex := spanIndex

	currentRegion := 0

	span := chf.spans[spanIndex]
	neighborSpanIndex := span.neighbors[dir]
	if neighborSpanIndex != -1 {
		currentRegion = regionIDs[neighborSpanIndex]
	}

	neighboringRegionIDs := []int{currentRegion}

	iter := 0
	for iter < 40000 {
		span := chf.spans[spanIndex]
		regionID := -1
		if isSolidEdge(chf, regionIDs, spanIndex, dir) {
			neighborSpanIndex := span.neighbors[dir]
			if neighborSpanIndex != -1 {
				// neighborSpan := chf.spans[neighborSpanIndex]
				regionID = regionIDs[neighborSpanIndex]
				if regionID != currentRegion {
					currentRegion = regionID
					neighboringRegionIDs = append(neighboringRegionIDs, regionID)
				}
			}
			// rotate CW
			dir = (dir + 1) % 4
		} else {
			spanIndex = span.neighbors[dir]
			if spanIndex == -1 {
				// this index should always be valid since we'd otherwise have found a solid edge
				panic("unexpected invalid index")
			}

			// rotate CCW
			dir = (dir + 3) % 4
		}

		if startIndex == spanIndex && startDir == dir {
			break
		}

		iter++
	}

	if len(neighboringRegionIDs) > 1 {
		for i := 0; i < len(neighboringRegionIDs); {
			j := (i + 1) % len(neighboringRegionIDs)
			if neighboringRegionIDs[i] == neighboringRegionIDs[j] {
				for k := i; k < len(neighboringRegionIDs)-1; k++ {
					neighboringRegionIDs[k] = neighboringRegionIDs[k+1]
				}
				neighboringRegionIDs = neighboringRegionIDs[:len(neighboringRegionIDs)-1]
			} else {
				i += 1
			}
		}

	}

	return neighboringRegionIDs
}

func isSolidEdge(chf *CompactHeightField, regionIDs []int, spanIndex SpanIndex, dir int) bool {
	span := chf.spans[spanIndex]

	neighborSpanIndex := span.neighbors[dir]
	if neighborSpanIndex == -1 {
		return true
	}

	if regionIDs[spanIndex] == regionIDs[neighborSpanIndex] {
		return false
	}
	return true
}

func addUniqueFloorRegion(region *Region, floorNum int) {
	for _, floor := range region.floors {
		if floor == floorNum {
			return
		}
	}
	region.floors = append(region.floors, floorNum)
}

type Region struct {
	spanCount        int
	id               int
	areaType         AREA_TYPE
	remap            bool
	visited          bool
	overlap          bool
	connectsToBorder bool
	yMin, yMax       int
	connections      []int
	floors           []int
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
