package navmesh

import (
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
	return &NavigationMesh{
		Volume:         collider.BoundingBox{MinVertex: mgl64.Vec3{-50, -50, -50}, MaxVertex: mgl64.Vec3{50, 50, 50}},
		voxelDimension: 1.0,
		world:          world,
	}
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
	var runs [3]int = [3]int{int(delta[0] / n.voxelDimension), int(delta[1] / n.voxelDimension), int(delta[2] / n.voxelDimension)}

	inputWorkCount := runs[0] * runs[1] * runs[2]
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

	n.voxelField = make([][][]Voxel, runs[0])
	for i := range n.voxelField {
		n.voxelField[i] = make([][]Voxel, runs[1])
		for j := range n.voxelField[i] {
			n.voxelField[i][j] = make([]Voxel, runs[2])
		}
	}

	for i := 0; i < runs[0]; i++ {
		for j := 0; j < runs[1]; j++ {
			for k := 0; k < runs[2]; k++ {
				inputWork <- [3]int{i, j, k}
			}
		}
	}
	close(inputWork)

	for work := range outputWork {
		n.voxelCount++

		x, y, z := work.x, work.y, work.z
		n.voxelField[x][y][z] = NewVoxel(x, y, z)
		n.voxelField[x][y][z].Filled = true
	}
	fmt.Printf("generated %d voxels\n", n.voxelCount)

	totalTricount := 0
	for _, count := range entityTriCount {
		totalTricount += count
	}
	fmt.Println("nav mesh entity tri count", totalTricount)
	return n.voxelField
}

func (n *NavigationMesh) buildNavigableArea(voxelField [][][]Voxel) [][][]Voxel {
	delta := n.Volume.MaxVertex.Sub(n.Volume.MinVertex)
	var runs [3]int = [3]int{int(delta[0] / n.voxelDimension), int(delta[1] / n.voxelDimension), int(delta[2] / n.voxelDimension)}

	work := make(chan [2]int, runs[0]*runs[2])

	workerCount := 12

	var wg sync.WaitGroup
	wg.Add(workerCount)

	// remove voxels that have another voxel directly above it
	for wc := 0; wc < workerCount; wc++ {
		go func() {
			defer wg.Done()
			for w := range work {
				for y := 0; y < runs[1]-1; y++ {
					x, z := w[0], w[1]
					if voxelField[x][y][z].Filled && voxelField[x][y+1][z].Filled {
						voxelField[x][y][z].Filled = false
					}
				}
			}
		}()
	}

	for i := 0; i < runs[0]; i++ {
		for j := 0; j < runs[2]; j++ {
			work <- [2]int{i, j}
		}
	}

	close(work)
	wg.Wait()

	return voxelField
}

func (n *NavigationMesh) computeDistanceTransform(voxelField [][][]Voxel) [][][]int {
	delta := n.Volume.MaxVertex.Sub(n.Volume.MinVertex)
	var runs [3]int = [3]int{int(delta[0] / n.voxelDimension), int(delta[1] / n.voxelDimension), int(delta[2] / n.voxelDimension)}

	// compute distance transform from the bottom up
	for y := 0; y < runs[1]; y++ {
		for x := 0; x < runs[0]; x++ {
			voxelField[x][y][0].DistanceField = 0
			voxelField[x][y][runs[2]-1].DistanceField = 0
		}
		for z := 0; z < runs[2]; z++ {
			voxelField[0][y][z].DistanceField = 0
			voxelField[runs[0]-1][y][z].DistanceField = 0
		}
	}

	for y := 0; y < runs[1]; y++ {
		for x := 0; x < runs[0]; x++ {
			for z := 0; z < runs[2]; z++ {
			}
		}
	}

	return nil
}

func (n *NavigationMesh) BakeNavMesh() {
	n.voxelField = n.voxelize()
	n.voxelField = n.buildNavigableArea(n.voxelField)
	distanceTransform := n.computeDistanceTransform(n.voxelField)
	_ = distanceTransform
}

type Voxel struct {
	Filled        bool
	X, Y, Z       int
	DistanceField int
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
		DistanceField: 999999999999999999,
	}
}
