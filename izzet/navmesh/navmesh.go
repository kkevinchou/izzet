package navmesh

import (
	"container/heap"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/kitolib/collision"
	"github.com/kkevinchou/kitolib/collision/collider"
	"github.com/kkevinchou/kitolib/spatialpartition"
	"github.com/kkevinchou/kitolib/utils"
)

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
		Volume:         collider.BoundingBox{MinVertex: mgl64.Vec3{-150, -25, -150}, MaxVertex: mgl64.Vec3{0, 150, 0}},
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

					for _, rd := range entity.Model.RenderData() {
						for _, tri := range meshTriangles[rd.MeshID] {
							if IntersectAABBTriangle(voxelAABB, tri) {
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
					// if voxelField[x][y][z].Filled && voxelField[x][y+1][z].Filled {
					// 	voxelField[x][y][z].Filled = false
					// }

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

func computeDistanceTransform(voxelField [][][]Voxel, dimensions [3]int) [][][]ReachInfo {
	reachField := computeReachField(voxelField, dimensions)

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

	// compute distance transform from the bottom up
	// we can skip over the boundary voxels since they're initialized above

	for x := 1; x < dimensions[0]-1; x++ {
		for z := 1; z < dimensions[2]-1; z++ {
			for y := 0; y < dimensions[1]; y++ {
				xNeighbor := voxelField[x-1][y][z]
				zNeighbor := voxelField[x][y][z-1]
				xzNegativeNeighbor := voxelField[x-1][y][z-1]
				xzPositiveNeighbor := voxelField[x+1][y][z-1]

				xReach := reachField[x-1][y][z]
				zReach := reachField[x][y][z-1]
				xzNegativeReach := reachField[x-1][y][z-1]
				xzPositiveReach := reachField[x+1][y][z-1]

				var minDistanceFieldValue float64

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

					minDistanceFieldValue = distanceFieldXValue
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
		}
	}

	for x := dimensions[0] - 2; x > 0; x-- {
		for z := dimensions[2] - 2; z > 0; z-- {
			for y := 0; y < dimensions[1]; y++ {
				xNeighbor := voxelField[x+1][y][z]
				zNeighbor := voxelField[x][y][z+1]
				xzNegativeNeighbor := voxelField[x+1][y][z+1]
				xzPositiveNeighbor := voxelField[x-1][y][z+1]

				xReach := reachField[x+1][y][z]
				zReach := reachField[x][y][z+1]
				xzNegativeReach := reachField[x+1][y][z+1]
				xzPositiveReach := reachField[x-1][y][z+1]

				var minDistanceFieldValue float64

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

					minDistanceFieldValue = distanceFieldXValue
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
		}
	}

	return reachField
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

func watershed(voxelField [][][]Voxel, reachField [][][]ReachInfo, dimensions [3]int) {
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

	// runCount := 0
	for pq.Len() > 0 {
		item := heap.Pop(&pq).(*Item)
		voxel := item.value

		x := voxel.X
		y := voxel.Y
		z := voxel.Z

		var nearestNeighbor *Voxel
		neighbors := getNeighbors(x, y, z, voxelField, reachField, dimensions)
		for _, neighbor := range neighbors {
			if nearestNeighbor == nil {
				nearestNeighbor = neighbor
			} else if neighbor.DistanceField > nearestNeighbor.DistanceField {
				nearestNeighbor = neighbor
			}
		}

		if nearestNeighbor != nil {
			// seed point since the distance is larger than any neighbor
			if voxel.DistanceField > nearestNeighbor.DistanceField || mgl64.FloatEqualThreshold(voxel.DistanceField, nearestNeighbor.DistanceField, 0.00001) {
				regionIDCounter++
				voxel.RegionID = regionIDCounter
				fmt.Println(regionIDCounter, "SEED -", voxel.X, voxel.Y, voxel.Z, voxel.DistanceField)
			} else {
				voxel.RegionID = nearestNeighbor.RegionID
			}
		} else {
			// isolated voxel, so it's its own region, though we'll probably discard it
			regionIDCounter++
			voxel.RegionID = regionIDCounter
			fmt.Println(regionIDCounter, "SEED -", voxel.X, voxel.Y, voxel.Z, voxel.DistanceField)
		}

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
	fmt.Println("DONE")
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

		if neighbor != nil {
			neighbors = append(neighbors, neighbor)
		}
	}

	return neighbors
}

func blurDistanceField(voxelField [][][]Voxel, reachField [][][]ReachInfo, dimensions [3]int) {
	for x := 0; x < dimensions[0]; x++ {
		for y := 0; y < dimensions[1]; y++ {
			for z := 0; z < dimensions[2]; z++ {
			}
		}
	}
}

func (n *NavigationMesh) BakeNavMesh() {
	delta := n.Volume.MaxVertex.Sub(n.Volume.MinVertex)
	var dimensions [3]int = [3]int{int(delta[0] / n.voxelDimension), int(delta[1] / n.voxelDimension), int(delta[2] / n.voxelDimension)}

	n.voxelField = n.voxelize()
	buildNavigableArea(n.voxelField, dimensions)
	reachField := computeDistanceTransform(n.voxelField, dimensions)
	watershed(n.voxelField, reachField, dimensions)
}

type Voxel struct {
	Filled        bool
	X, Y, Z       int
	DistanceField float64
	Seed          bool
	RegionID      int

	// Neighbors [][3]int
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
