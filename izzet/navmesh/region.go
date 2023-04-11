package navmesh

import (
	"container/heap"
	"fmt"
	"sort"

	"github.com/go-gl/mathgl/mgl32"
)

var initialBorderCell map[int]*Voxel

func watershed(voxelField [][][]Voxel, reachField [][][]ReachInfo, dimensions [3]int) map[int][]VoxelPosition {
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

	regionMap := map[int][]VoxelPosition{}

	// runCount := 0
	for pq.Len() > 0 {
		item := heap.Pop(&pq).(*Item)
		voxel := item.value

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

		regionMap[voxel.RegionID] = append(regionMap[voxel.RegionID], VoxelPosition{x, y, z})
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

func traceRegionContours(voxelField [][][]Voxel, reachField [][][]ReachInfo, dimensions [3]int, regionMap map[int][]VoxelPosition, initialBorderVoxel map[int]*Voxel) {
	for _, voxel := range initialBorderVoxel {
		traceRegionContour(voxelField, reachField, dimensions, regionMap, voxel)
	}
}

type Contour struct {
}

var movementRight [2]int = [2]int{1, 0}
var movementLeft [2]int = [2]int{-1, 0}
var movementUp [2]int = [2]int{0, -1}
var movementDown [2]int = [2]int{0, 1}

var movementRightUp [2]int = [2]int{1, -1}
var movementRightDown [2]int = [2]int{1, 1}
var movementLeftUp [2]int = [2]int{-1, -1}
var movementLeftDown [2]int = [2]int{-1, 1}

// prefer up/down/left/right over diagonals when navigating contours
var traceNeighborDirs [][2]int
var movementToNeighborDir map[[2]int]int

func init() {
	traceNeighborDirs = [][2]int{
		// [2]int{0, -1}, [2]int{0, 1}, [2]int{-1, 0}, [2]int{1, 0},
		// [2]int{-1, -1}, [2]int{1, -1}, [2]int{-1, 1}, [2]int{1, 1},

		// start with up
		[2]int{0, -1},
		// [2]int{-1, -1},
		[2]int{-1, 0},
		// [2]int{-1, 1},
		[2]int{0, 1},
		// [2]int{1, 1},
		[2]int{1, 0},
		// [2]int{1, -1},
	}

	movementToNeighborDir = map[[2]int]int{}
	for i, dir := range traceNeighborDirs {
		movementToNeighborDir[dir] = i
	}
}

// TODO - gonna need to decide on how I want to handle holes
func traceRegionContour(voxelField [][][]Voxel, reachField [][][]ReachInfo, dimensions [3]int, regionMap map[int][]VoxelPosition, startVoxel *Voxel) Contour {
	var next *Voxel
	seen := map[VoxelPosition]bool{}

	// initialization
	neighbors := getNeighborsOrdered(startVoxel.X, startVoxel.Y, startVoxel.Z, voxelField, reachField, dimensions, traceNeighborDirs, 0)
	for _, neighbor := range neighbors {
		if !neighbor.Border {
			continue
		}

		if neighbor.RegionID == startVoxel.RegionID {
			next = neighbor
		}
	}

	if next == nil {
		return Contour{}
	}
	firstNextVoxel := next

	secondMovement := [2]int{next.X - startVoxel.X, next.Z - startVoxel.Z}
	var firstMovement [2]int

	for next != nil {
		voxel := next
		next = nil
		seen[voxelPos(voxel)] = true

		if startVoxel.RegionID == 1 && voxel.X == 94 && voxel.Z == 118 {
			fmt.Println(*voxel)
		}

		neighbors := getNeighborsOrdered(voxel.X, voxel.Y, voxel.Z, voxelField, reachField, dimensions, traceNeighborDirs, movementToNeighborDir[secondMovement])
		for _, neighbor := range neighbors {
			if !neighbor.Border || seen[voxelPos(neighbor)] {
				continue
			}

			if neighbor.RegionID == voxel.RegionID {
				next = neighbor
				break
			}
		}

		if next != nil {
			firstMovement = secondMovement
			secondMovement = [2]int{next.X - voxel.X, next.Z - voxel.Z}
			if firstMovement != secondMovement {
				voxel.ContourCorner = true
				var c float32 = 4
				voxel.DEBUGCOLORFACTOR = &c
			}
		}
	}

	// handle the edge case where the start voxel is a corner voxel
	firstMovement = secondMovement
	secondMovement = [2]int{firstNextVoxel.X - startVoxel.X, firstNextVoxel.Z - startVoxel.Z}
	if firstMovement != secondMovement {
		startVoxel.ContourCorner = true
		var c float32 = 4
		startVoxel.DEBUGCOLORFACTOR = &c
	}

	return Contour{}
}

// TODO - update this to use half edges to collect voxels
func mergeRegions(voxelField [][][]Voxel, reachField [][][]ReachInfo, dimensions [3]int, regionMap map[int][]VoxelPosition) {
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

			if seen[voxelPos(voxel)] {
				continue
			}

			seen[voxelPos(voxel)] = true
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
				if seen[voxelPos(neighbor)] {
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
			regionMap[nextRegionID] = append(regionMap[nextRegionID], voxelPos(&voxelField[x][y][z]))
		}
		delete(regionMap, regionID)
	}
}

var borderNeighborDirs [][2]int = [][2]int{
	[2]int{-1, -1}, [2]int{0, -1}, [2]int{1, -1},
	[2]int{1, 0},
	[2]int{1, 1}, [2]int{0, 1}, [2]int{-1, 1},
	[2]int{-1, 0},
}

func markBorderVoxels(voxelField [][][]Voxel, reachField [][][]ReachInfo, dimensions [3]int, regionMap map[int][]VoxelPosition) map[int]*Voxel {
	borderVoxel := map[int]*Voxel{}
	for x := 0; x < dimensions[0]; x++ {
		for y := 0; y < dimensions[1]; y++ {
			for z := 0; z < dimensions[2]; z++ {
				voxel := &voxelField[x][y][z]
				if voxel.RegionID == 1 && x == 46 && z == 40 {
					fmt.Println("HI")
				}
				neighbors := getNeighborsOrdered(voxel.X, voxel.Y, voxel.Z, voxelField, reachField, dimensions, borderNeighborDirs, 0)
				if len(neighbors) != 8 {
					voxel.Border = true
				} else {
					if voxel.X == 52 && voxel.Y == 76 && voxel.Z == 6 {
						fmt.Println(*voxel)
					}
					// there is an edge case where a voxel is surrounded by voxels of the same region but is still a border voxel.
					// this happens when one the neighbor voxels is not reachable by one of it's neighboring voxels (e.g. left or right).
					// in that scenario, our current voxel is a border voxel since we would need to travel through this voxel to connect
					// those two neighboring voxels
					for i := 0; i < len(neighbors); i++ {
						if voxel.RegionID != neighbors[i].RegionID {
							voxel.Border = true
							break
						}
						if !isConnected(neighbors[i], neighbors[(i+1)%len(neighbors)], voxelField, reachField) {
							voxel.Border = true
							break
						}
					}
				}
				if voxel.Border {
					if prev, ok := borderVoxel[voxel.RegionID]; ok {
						if voxel.X > prev.X {
							borderVoxel[voxel.RegionID] = voxel
						} else if voxel.X == prev.X {
							if voxel.Z > prev.Z {
								borderVoxel[voxel.RegionID] = voxel
							}
						}
					} else {
						borderVoxel[voxel.RegionID] = voxel
					}
					// var c float32 = 0.5
					// voxel.DEBUGCOLORFACTOR = &c
				}
			}
		}
	}
	return borderVoxel
}

func filterRegions(voxelField [][][]Voxel, reachField [][][]ReachInfo, dimensions [3]int, regionMap map[int][]VoxelPosition) {
	// filter regions that are too small and isolated
	// might require region definition first
}

func neighborDist(voxel, neighbor *Voxel) float64 {
	dist := neighbor.DistanceField + 1
	if isDiagonal(voxel, neighbor) {
		// diagonals are slightly further
		dist += 0.4
	}
	return dist
}

func isDiagonal(a, b *Voxel) bool {
	return a.X != b.X && a.Z != b.Z
}

var regionIDCounter int
