package navmesh

import (
	"container/heap"
	"fmt"
	"sort"

	"github.com/go-gl/mathgl/mgl32"
)

func watershed(voxelField [][][]Voxel, reachField [][][]ReachInfo, dimensions [3]int) map[int][][3]int {
	pq := PriorityQueue{}

	for x := 0; x < dimensions[0]; x++ {
		for y := 0; y < dimensions[1]; y++ {
			for z := 0; z < dimensions[2]; z++ {
				if !voxelField[x][y][z].Filled {
					continue
				}

				voxel := &voxelField[x][y][z]
				pq = append(pq, &Item{value: voxel, priority: voxel.DistanceField})
			}
		}
	}
	heap.Init(&pq)

	regionMap := map[int][][3]int{}
	priorityMap := map[float64]int{}

	// runCount := 0
	for pq.Len() > 0 {
		item := heap.Pop(&pq).(*Item)
		voxel := item.value
		priorityMap[item.priority]++

		x := voxel.X
		y := voxel.Y
		z := voxel.Z

		isSeed := true
		var nearestNeighbor *Voxel

		neighbors := getNeighbors(x, y, z, voxelField, reachField, dimensions)
		for _, neighbor := range neighbors {
			if neighbor.RegionID == -1 {
				continue
			}
			if mgl32.FloatEqualThreshold(float32(voxel.DistanceField), float32(neighbor.DistanceField), 0.00000001) {
				// voxel is from the same pass, don't consider neighbors that happened to be processed at the same time
				continue
			}

			if nearestNeighbor == nil {
				nearestNeighbor = neighbor
			} else if neighborDist(voxel, neighbor) < neighborDist(voxel, nearestNeighbor) {
				nearestNeighbor = neighbor
			}
			if neighbor.DistanceField > voxel.DistanceField {
				isSeed = false
			}
		}

		if isSeed {
			regionIDCounter++
			voxel.RegionID = regionIDCounter
			voxel.Seed = true
		} else if nearestNeighbor != nil {
			voxel.RegionID = nearestNeighbor.RegionID
		}

		regionMap[voxel.RegionID] = append(regionMap[voxel.RegionID], [3]int{x, y, z})
	}

	regionCount := map[int]int{}
	for x := 0; x < dimensions[0]; x++ {
		for y := 0; y < dimensions[1]; y++ {
			for z := 0; z < dimensions[2]; z++ {
				if !voxelField[x][y][z].Filled {
					continue
				}

				voxel := &voxelField[x][y][z]
				regionCount[voxel.RegionID]++
			}
		}
	}
	fmt.Printf("generated %d regions\n", len(regionCount))
	return regionMap
}

func mergeRegions(voxelField [][][]Voxel, reachField [][][]ReachInfo, dimensions [3]int, regionMap map[int][][3]int) {
	processedRegions := map[int]bool{}
	regionConversion := map[int]int{}

	var regionIDs []int
	for id := range regionMap {
		regionIDs = append(regionIDs, id)
	}
	sort.Ints(regionIDs)

	for _, regionID := range regionIDs {
		if _, ok := processedRegions[regionID]; ok {
			continue
		}

		// the first coordinates always belongs to the seed voxel
		coords := regionMap[regionID][0]
		x := coords[0]
		y := coords[1]
		z := coords[2]

		var search []*Voxel
		search = append(search, &voxelField[x][y][z])
		seen := map[[3]int]bool{}

		for len(search) > 0 {
			voxel := search[0]
			search = search[1:]

			if seen[[3]int{voxel.X, voxel.Y, voxel.Z}] {
				continue
			}

			seen[[3]int{voxel.X, voxel.Y, voxel.Z}] = true
			processedRegions[voxel.RegionID] = true
			regionConversion[voxel.RegionID] = regionID

			neighbors := getNeighbors(voxel.X, voxel.Y, voxel.Z, voxelField, reachField, dimensions)
			for _, neighbor := range neighbors {
				if !neighbor.Seed {
					continue
				}
				if _, ok := processedRegions[neighbor.RegionID]; ok {
					continue
				}
				if seen[[3]int{neighbor.X, neighbor.Y, neighbor.Z}] {
					continue
				}

				search = append(search, neighbor)
			}
		}
	}

	for _, regionID := range regionIDs {
		if _, ok := regionConversion[regionID]; !ok {
			continue
		}

		nextRegionID := regionConversion[regionID]
		if regionID == nextRegionID {
			continue
		}

		coords := regionMap[regionID]
		for _, coord := range coords {
			x := coord[0]
			y := coord[1]
			z := coord[2]
			voxelField[x][y][z].RegionID = nextRegionID
		}
	}
}
