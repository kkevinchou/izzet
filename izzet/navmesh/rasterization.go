package navmesh

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/kitolib/collision"
	"github.com/kkevinchou/kitolib/collision/collider"
	"github.com/kkevinchou/kitolib/utils"
)

type MinXTriangles []Triangle

func (u MinXTriangles) Len() int {
	return len(u)
}
func (u MinXTriangles) Swap(i, j int) {
	u[i], u[j] = u[j], u[i]
}
func (u MinXTriangles) Less(i, j int) bool {
	return u[i].MinX < u[j].MinX
}

func (n *NavigationMesh) voxelize() [][][]Voxel {
	start := time.Now()
	spatialPartition := n.world.SpatialPartition()
	sEntities := spatialPartition.QueryEntities(n.Volume)

	var candidateEntities []*entities.Entity
	boundingBoxes := map[int]collider.BoundingBox{}
	meshTriangles := map[int][]Triangle{}
	entityTriCount := map[int]int{}

	// collect candidate entities for voxelization
	for _, entity := range sEntities {
		e := n.world.GetEntityByID(entity.GetID())
		if e == nil {
			continue
		}

		var shouldProcessEntity bool

		if strings.Contains(e.Model.Name(), "Tile") || strings.Contains(e.Model.Name(), "Stair") {
			shouldProcessEntity = true
		} else if e.Parent != nil && strings.Contains(e.Parent.Name, "simple_plane") {
			shouldProcessEntity = true
		}

		if !shouldProcessEntity {
			continue
		}

		candidateEntities = append(candidateEntities, e)
		boundingBoxes[e.GetID()] = *e.BoundingBox()
		transform := utils.Mat4F64ToF32(entities.WorldTransform(e))

		for _, rd := range e.Model.RenderData() {
			meshID := rd.MeshID
			mesh := e.Model.Collection().Meshes[meshID]
			for i := 0; i < len(mesh.Vertices); i += 3 {
				t := Triangle{
					V1: utils.Vec3F32ToF64(transform.Mul4x1(mesh.Vertices[i].Position.Vec4(1)).Vec3()),
					V2: utils.Vec3F32ToF64(transform.Mul4x1(mesh.Vertices[i+1].Position.Vec4(1)).Vec3()),
					V3: utils.Vec3F32ToF64(transform.Mul4x1(mesh.Vertices[i+2].Position.Vec4(1)).Vec3()),
				}
				t.MinX = math.Min(t.V1.X(), math.Min(t.V2.X(), t.V3.X()))
				meshTriangles[meshID] = append(meshTriangles[meshID], t)
			}
			numVerts := len(mesh.Vertices)
			entityTriCount[e.GetID()] = numVerts / 3
		}
	}

	// for more complex geometry it may be worth actually creating an
	// oct tree for the mesh. realistically we shouldn't be using
	// very complicated geometry for generating nav meshes
	for _, triangles := range meshTriangles {
		sort.Sort(MinXTriangles(triangles))
	}

	outputWork := make(chan OutputWork)
	delta := n.Volume.MaxVertex.Sub(n.Volume.MinVertex)
	var dimensions [3]int = [3]int{int(delta[0] / n.voxelDimension), int(delta[1] / n.voxelDimension), int(delta[2] / n.voxelDimension)}

	inputWorkCount := dimensions[0] * dimensions[1] * dimensions[2]
	inputWork := make(chan VoxelPosition, inputWorkCount)
	workerCount := 12

	doneWorkerCount := 0
	var doneWorkerMutex sync.Mutex

	// set up workers perform voxelization at a specific 3d coordinate
	for i := 0; i < workerCount; i++ {
		go func() {
			for input := range inputWork {
				x, y, z := input[0], input[1], input[2]

				voxel := &collider.BoundingBox{
					MinVertex: n.Volume.MinVertex.Add(mgl64.Vec3{float64(x), float64(y), float64(z)}.Mul(n.voxelDimension)),
					MaxVertex: n.Volume.MinVertex.Add(mgl64.Vec3{float64(x + 1), float64(y + 1), float64(z + 1)}.Mul(n.voxelDimension)),
				}
				voxelAABB := AABB{Min: voxel.MinVertex, Max: voxel.MaxVertex}

				for _, entity := range candidateEntities {
					bb := boundingBoxes[entity.GetID()]
					if !collision.CheckOverlapAABBAABB(voxel, &bb) {
						continue
					}

					for _, rd := range entity.Model.RenderData() {
						for _, tri := range meshTriangles[rd.MeshID] {
							// NOTE - rather than doing an expensive AABB/Triangle intersection
							// Recast clips the triangle against the voxels in the heighfield.
							// that implementation is likely a lot more performant
							if IntersectAABBTriangle(voxelAABB, tri) {
								outputWork <- OutputWork{
									x:           x,
									y:           y,
									z:           z,
									boundingBox: *voxel,
								}

								goto Done
							}
							if voxelAABB.Max.X() < tri.MinX {
								continue
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

	// initialize the voxel field
	voxelField := make([][][]Voxel, dimensions[0])
	for i := range voxelField {
		voxelField[i] = make([][]Voxel, dimensions[1])
		for j := range voxelField[i] {
			voxelField[i][j] = make([]Voxel, dimensions[2])
		}
	}

	// create a work item for each voxel location
	for i := 0; i < dimensions[0]; i++ {
		for j := 0; j < dimensions[1]; j++ {
			for k := 0; k < dimensions[2]; k++ {
				inputWork <- VoxelPosition{i, j, k}
			}
		}
	}
	close(inputWork)

	// assemble voxels into the voxel field
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

	// remove voxels that are too low - i.e. there is a voxel from above that
	// would interfere with the agent height
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

func fillHoles(voxelField [][][]Voxel, dimensions [3]int) {
	for y := 0; y < dimensions[1]; y++ {
		for z := 0; z < dimensions[2]; z++ {
			for x := 0; x < dimensions[0]; x++ {
				if voxelField[x][y][z].Filled {
					continue
				}

				// voxel := &voxelField[x][y][z]
			}
		}
	}
}

type Voxel struct {
	Filled           bool
	X, Y, Z          int
	DistanceField    float64
	Seed             bool
	RegionID         int
	DEBUGCOLORFACTOR *float32
	Border           bool
	ContourCorner    bool
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
