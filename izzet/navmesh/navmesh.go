package navmesh

import (
	"container/heap"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/kitolib/collision"
	"github.com/kkevinchou/kitolib/collision/collider"
	"github.com/kkevinchou/kitolib/spatialpartition"
	"github.com/kkevinchou/kitolib/utils"
)

// GENERAL IMPLEMENTATION NOTES

// in general the generated regions in a nav mesh should be free of holes.
// due to the resolution of the voxels there are degenerate cases where holes
// can be present in the generated mesh regions. for example, holes in meshes
// with a size of 1 or 2 tend to be ignored. however, larger holes will be properly
// processed

const stepHeight int = 4
const agentHeight int = 30

type World interface {
	SpatialPartition() *spatialpartition.SpatialPartition
	GetEntityByID(id int) *entities.Entity
}

type NavigationMesh struct {
	Volume collider.BoundingBox
	world  World

	voxelCount     int
	voxelField     [][][]Voxel
	voxelDimension float64
}

func New(world World) *NavigationMesh {
	nm := &NavigationMesh{
		// Volume: collider.BoundingBox{MinVertex: mgl64.Vec3{75, -50, -200}, MaxVertex: mgl64.Vec3{350, 25, -50}},
		// Volume: collider.BoundingBox{MinVertex: mgl64.Vec3{-150, -50, -350}, MaxVertex: mgl64.Vec3{350, 150, 150}},
		Volume: collider.BoundingBox{MinVertex: mgl64.Vec3{-150, -25, -150}, MaxVertex: mgl64.Vec3{150, 150, 0}},
		// Volume: collider.BoundingBox{MinVertex: mgl64.Vec3{0, -25, 0}, MaxVertex: mgl64.Vec3{100, 100, 150}},
		// Volume:         collider.BoundingBox{MinVertex: mgl64.Vec3{-50, -25, 0}, MaxVertex: mgl64.Vec3{100, 100, 150}},
		voxelDimension: 1.0,
		world:          world,
	}
	nm.BakeNavMesh()
	// move the scene out of the way
	entity := world.GetEntityByID(3)
	entities.SetLocalPosition(entity, mgl64.Vec3{0, -1000, 0})
	return nm
}

func (n *NavigationMesh) voxelize() [][][]Voxel {
	start := time.Now()
	spatialPartition := n.world.SpatialPartition()
	sEntities := spatialPartition.QueryEntities(n.Volume)

	var candidateEntities []*entities.Entity
	boundingBoxes := map[int]collider.BoundingBox{}
	meshTriangles := map[int][]Triangle{}
	entityTriCount := map[int]int{}

	for _, entity := range sEntities {
		e := n.world.GetEntityByID(entity.GetID())
		if e == nil {
			continue
		}

		if !strings.Contains(e.Model.Name(), "Tile") && !strings.Contains(e.Model.Name(), "Stair") {
			continue
		}

		candidateEntities = append(candidateEntities, e)
		boundingBoxes[e.GetID()] = *e.BoundingBox()
		transform := utils.Mat4F64ToF32(entities.WorldTransform(e))

		for _, rd := range e.Model.RenderData() {
			meshID := rd.MeshID
			mesh := e.Model.Collection().Meshes[meshID]
			for i := 0; i < len(mesh.Vertices); i += 3 {
				meshTriangles[meshID] = append(meshTriangles[meshID],
					Triangle{
						V1: utils.Vec3F32ToF64(transform.Mul4x1(mesh.Vertices[i].Position.Vec4(1)).Vec3()),
						V2: utils.Vec3F32ToF64(transform.Mul4x1(mesh.Vertices[i+1].Position.Vec4(1)).Vec3()),
						V3: utils.Vec3F32ToF64(transform.Mul4x1(mesh.Vertices[i+2].Position.Vec4(1)).Vec3()),
					},
				)
			}
			numVerts := len(mesh.Vertices)
			entityTriCount[e.GetID()] = numVerts / 3
		}
	}

	outputWork := make(chan OutputWork)
	delta := n.Volume.MaxVertex.Sub(n.Volume.MinVertex)
	var dimensions [3]int = [3]int{int(delta[0] / n.voxelDimension), int(delta[1] / n.voxelDimension), int(delta[2] / n.voxelDimension)}

	inputWorkCount := dimensions[0] * dimensions[1] * dimensions[2]
	inputWork := make(chan [3]int, inputWorkCount)
	workerCount := 12

	doneWorkerCount := 0
	var doneWorkerMutex sync.Mutex

	for wi := 0; wi < workerCount; wi++ {
		go func() {
			for input := range inputWork {
				i := input[0]
				j := input[1]
				k := input[2]

				voxel := &collider.BoundingBox{
					MinVertex: n.Volume.MinVertex.Add(mgl64.Vec3{float64(i), float64(j), float64(k)}.Mul(n.voxelDimension)),
					MaxVertex: n.Volume.MinVertex.Add(mgl64.Vec3{float64(i + 1), float64(j + 1), float64(k + 1)}.Mul(n.voxelDimension)),
				}
				voxelAABB := AABB{Min: voxel.MinVertex, Max: voxel.MaxVertex}

				for _, entity := range candidateEntities {
					bb := boundingBoxes[entity.GetID()]
					if !collision.CheckOverlapAABBAABB(voxel, &bb) {
						continue
					}

					// 1330 is a tile polygon with almost no y thickness
					id := entity.GetID()
					if id == 1330 {
						// fmt.Println("A")
					}

					for _, rd := range entity.Model.RenderData() {
						for _, tri := range meshTriangles[rd.MeshID] {
							if IntersectAABBTriangle(voxelAABB, tri) {
								id := entity.GetID()
								if id == 1330 {
									// fmt.Println("B")
								}
								outputWork <- OutputWork{
									x:           i,
									y:           j,
									z:           k,
									boundingBox: *voxel,
								}

								goto Done
							}
						}
					}
				Done:
				}
			}

			doneWorkerMutex.Lock()
			doneWorkerCount++
			if doneWorkerCount == workerCount {
				fmt.Println("generation time seconds", time.Since(start).Seconds())
				close(outputWork)
			}
			doneWorkerMutex.Unlock()
		}()
	}

	voxelField := make([][][]Voxel, dimensions[0])
	for i := range voxelField {
		voxelField[i] = make([][]Voxel, dimensions[1])
		for j := range voxelField[i] {
			voxelField[i][j] = make([]Voxel, dimensions[2])
		}
	}

	for i := 0; i < dimensions[0]; i++ {
		for j := 0; j < dimensions[1]; j++ {
			for k := 0; k < dimensions[2]; k++ {
				inputWork <- [3]int{i, j, k}
			}
		}
	}
	close(inputWork)

	for work := range outputWork {
		n.voxelCount++

		x, y, z := work.x, work.y, work.z
		voxelField[x][y][z] = NewVoxel(x, y, z)
		voxelField[x][y][z].Filled = true
	}
	fmt.Printf("generated %d voxels\n", n.voxelCount)

	totalTricount := 0
	for _, count := range entityTriCount {
		totalTricount += count
	}
	fmt.Println("nav mesh entity tri count", totalTricount)
	return voxelField
}

func buildNavigableArea(voxelField [][][]Voxel, dimensions [3]int) {
	work := make(chan [2]int, dimensions[0]*dimensions[2])
	workerCount := 12

	var wg sync.WaitGroup
	wg.Add(workerCount)

	// remove voxels that have another voxel directly above it
	for wc := 0; wc < workerCount; wc++ {
		go func() {
			defer wg.Done()
			for w := range work {
				for y := dimensions[1] - 1; y >= 0; y-- {
					x, z := w[0], w[1]
					if voxelField[x][y][z].Filled {
						for i := 1; i < agentHeight+1; i++ {
							if y-i < 0 {
								break
							}
							voxelField[x][y-i][z] = NewVoxel(x, y-i, z)
						}
						y -= agentHeight
					}
				}
			}
		}()
	}

	for i := 0; i < dimensions[0]; i++ {
		for j := 0; j < dimensions[2]; j++ {
			work <- [2]int{i, j}
		}
	}

	close(work)
	wg.Wait()
}

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

type ReachInfo struct {
	sourceVoxel *Voxel
	source      [3]int
	hasSource   bool
}

func computeReachField(voxelField [][][]Voxel, dimensions [3]int) [][][]ReachInfo {
	reachField := make([][][]ReachInfo, dimensions[0])
	for i := range reachField {
		reachField[i] = make([][]ReachInfo, dimensions[1])
		for j := range reachField[i] {
			reachField[i][j] = make([]ReachInfo, dimensions[2])
		}
	}

	for y := 0; y < dimensions[1]-stepHeight; y++ {
		for x := 0; x < dimensions[0]; x++ {
			for z := 0; z < dimensions[2]; z++ {
				if !voxelField[x][y][z].Filled {
					continue
				}

				for i := 1; i < stepHeight+1; i++ {
					if y+i < dimensions[1] {
						reachField[x][y+i][z].hasSource = true
						reachField[x][y+i][z].source = [3]int{x, y, z}
						reachField[x][y+i][z].sourceVoxel = &voxelField[x][y][z]
					}

					if y-i >= 0 {
						reachField[x][y-i][z].hasSource = true
						reachField[x][y-i][z].source = [3]int{x, y, z}
						reachField[x][y-i][z].sourceVoxel = &voxelField[x][y][z]
					}
				}
			}
		}
	}

	return reachField
}

var neighborDirs [][2]int = [][2]int{
	[2]int{-1, -1}, [2]int{0, -1}, [2]int{1, -1},
	[2]int{-1, 0} /* current voxel */, [2]int{1, 0},
	[2]int{-1, 1}, [2]int{0, 1}, [2]int{1, 1},
}
var regionIDCounter int

func isDiagonal(a, b *Voxel) bool {
	return a.X != b.X && a.Z != b.Z
}

func neighborDist(voxel, neighbor *Voxel) float64 {
	dist := neighbor.DistanceField + 1
	if isDiagonal(voxel, neighbor) {
		// diagonals are slightly further
		dist += 0.4
	}
	return dist
}

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
		// var nearestDistance float64
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

		if nearestNeighbor != nil {
			// seed point since the distance is larger than any neighbor
			// if voxel.DistanceField > nearestNeighbor.DistanceField || mgl64.FloatEqualThreshold(voxel.DistanceField, nearestNeighbor.DistanceField, 0.00001) {
			if isSeed {
				regionIDCounter++
				voxel.RegionID = regionIDCounter
				voxel.Seed = true
				// fmt.Println("SEED - ", voxel.X, voxel.Y, voxel.Z, voxel.DistanceField)
			} else {
				voxel.RegionID = nearestNeighbor.RegionID
			}
		} else {
			// isolated voxel, so it's its own region, though we'll probably discard it
			regionIDCounter++
			voxel.Seed = true
			voxel.RegionID = regionIDCounter
		}

		regionMap[voxel.RegionID] = append(regionMap[voxel.RegionID], [3]int{x, y, z})

		// fmt.Println(voxel.X, voxel.Y, voxel.Z)

		// runCount++
		// if runCount >= 7 {
		// 	break
		// }
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

func getDebugNeighbors(x, y, z int, voxelField [][][]Voxel, reachField [][][]ReachInfo, dimensions [3]int) []*Voxel {
	var neighbors []*Voxel

	for _, dir := range neighborDirs {
		if x+dir[0] < 0 || z+dir[1] < 0 || x+dir[0] >= dimensions[0] || z+dir[1] >= dimensions[2] {
			neighbors = append(neighbors, nil)
			continue
		}

		var neighbor *Voxel
		reachNeighbor := &reachField[x+dir[0]][y][z+dir[1]]
		if reachNeighbor.sourceVoxel != nil {
			neighbor = reachNeighbor.sourceVoxel
		} else if voxelField[x+dir[0]][y][z+dir[1]].Filled {
			neighbor = &voxelField[x+dir[0]][y][z+dir[1]]
		}

		neighbors = append(neighbors, neighbor)
	}

	return neighbors
}

func getNeighbors(x, y, z int, voxelField [][][]Voxel, reachField [][][]ReachInfo, dimensions [3]int) []*Voxel {
	var neighbors []*Voxel

	for _, dir := range neighborDirs {
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

		neighbors = append(neighbors, neighbor)
	}

	return neighbors
}

//	var neighborDirs [][2]int = [][2]int{
//		[2]int{-1, -1}, [2]int{0, -1}, [2]int{1, -1},
//		[2]int{-1, 0} /* current voxel */, [2]int{1, 0},
//		[2]int{-1, 1}, [2]int{0, 1}, [2]int{1, 1},
//	}

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

	// fmt.Println("PRE BLUR ======================================")
	// printRegionMin := mgl32.Vec3{99, 25, 134}
	// printRegionMax := mgl32.Vec3{108, 25, 146}
	// for y := 0; y < dimensions[1]; y++ {
	// 	for z := 0; z < dimensions[2]; z++ {
	// 		for x := 0; x < dimensions[0]; x++ {
	// 			voxel := &voxelField[x][y][z]
	// 			if x >= int(printRegionMin.X()) && x <= int(printRegionMax.X()) {
	// 				if y >= int(printRegionMin.Y()) && y <= int(printRegionMax.Y()) {
	// 					if z >= int(printRegionMin.Z()) && z <= int(printRegionMax.Z()) {
	// 						if voxel.Filled {
	// 							fmt.Printf("[%5.3f] ", voxel.DistanceField)
	// 						} else {
	// 							fmt.Printf("[-----] ")
	// 						}
	// 						if voxel.X == int(printRegionMax.X()) {
	// 							fmt.Printf("\n")
	// 						}
	// 					}
	// 				}
	// 			}
	// 		}
	// 	}
	// }

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

	// fmt.Println("POST BLUR ======================================")
	// for y := 0; y < dimensions[1]; y++ {
	// 	for z := 0; z < dimensions[2]; z++ {
	// 		for x := 0; x < dimensions[0]; x++ {
	// 			voxel := &voxelField[x][y][z]
	// 			if x >= int(printRegionMin.X()) && x <= int(printRegionMax.X()) {
	// 				if y >= int(printRegionMin.Y()) && y <= int(printRegionMax.Y()) {
	// 					if z >= int(printRegionMin.Z()) && z <= int(printRegionMax.Z()) {
	// 						if voxel.Filled {
	// 							fmt.Printf("[%5.3f] ", voxel.DistanceField)
	// 						} else {
	// 							fmt.Printf("[-----] ")
	// 						}
	// 						if voxel.X == int(printRegionMax.X()) {
	// 							fmt.Printf("\n")
	// 						}
	// 					}
	// 				}
	// 			}
	// 		}
	// 	}
	// }

	// arrows := []string{
	// 	"🡔", "🡑", "🡕",
	// 	"←" /* center */, "🡒",
	// 	"🡗", "🡓", "🡖",
	// }
	// fmt.Println("POST BLUR ARROWS ======================================")
	// for y := 0; y < dimensions[1]; y++ {
	// 	for z := 0; z < dimensions[2]; z++ {
	// 		for x := 0; x < dimensions[0]; x++ {
	// 			voxel := &voxelField[x][y][z]
	// 			if x >= int(printRegionMin.X()) && x <= int(printRegionMax.X()) {
	// 				if y >= int(printRegionMin.Y()) && y <= int(printRegionMax.Y()) {
	// 					if z >= int(printRegionMin.Z()) && z <= int(printRegionMax.Z()) {
	// 						if voxel.Filled {
	// 							neighbors := getDebugNeighbors(x, y, z, voxelField, reachField, dimensions)
	// 							var maxNeighbor *Voxel

	// 							var arrow string
	// 							for index, neighbor := range neighbors {
	// 								if neighbor == nil {
	// 									continue
	// 								}
	// 								if maxNeighbor == nil {
	// 									maxNeighbor = neighbor
	// 									arrow = arrows[index]
	// 								} else {
	// 									if neighbor.DistanceField > maxNeighbor.DistanceField {
	// 										arrow = arrows[index]
	// 										maxNeighbor = neighbor
	// 									}
	// 								}
	// 							}
	// 							fmt.Printf("[%s] ", arrow)
	// 						} else {
	// 							fmt.Printf("[-] ")
	// 						}
	// 						if voxel.X == int(printRegionMax.X()) {
	// 							fmt.Printf("\n")
	// 						}

	// 					}
	// 				}
	// 			}
	// 		}
	// 	}
	// }
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

func (n *NavigationMesh) BakeNavMesh() {
	delta := n.Volume.MaxVertex.Sub(n.Volume.MinVertex)
	var dimensions [3]int = [3]int{int(delta[0] / n.voxelDimension), int(delta[1] / n.voxelDimension), int(delta[2] / n.voxelDimension)}

	n.voxelField = n.voxelize()
	buildNavigableArea(n.voxelField, dimensions)
	reachField := computeReachField(n.voxelField, dimensions)
	computeDistanceTransform(n.voxelField, reachField, dimensions)
	blurDistanceField(n.voxelField, reachField, dimensions)
	regionMap := watershed(n.voxelField, reachField, dimensions)
	_ = regionMap
	mergeRegions(n.voxelField, reachField, dimensions, regionMap)
}

type Voxel struct {
	Filled        bool
	X, Y, Z       int
	DistanceField float64
	Seed          bool
	RegionID      int
}

type OutputWork struct {
	x, y, z     int
	boundingBox collider.BoundingBox
}

func NewVoxel(x, y, z int) Voxel {
	return Voxel{
		Filled:        false,
		X:             x,
		Y:             y,
		Z:             z,
		DistanceField: MaxDistanceFieldValue,
		RegionID:      -1,
	}
}

const MaxDistanceFieldValue float64 = 999999999999999
