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

// var NavMesh *NavigationMesh = New()

type World interface {
	SpatialPartition() *spatialpartition.SpatialPartition
	GetEntityByID(id int) *entities.Entity
}

type NavigationMesh struct {
	Volume collider.BoundingBox
	world  World

	vertices []mgl64.Vec3
	normals  []mgl64.Vec3

	mutex sync.Mutex
}

func New(world World) *NavigationMesh {
	return &NavigationMesh{
		Volume: collider.BoundingBox{MinVertex: mgl64.Vec3{-50, -50, -50}, MaxVertex: mgl64.Vec3{50, 50, 50}},
		world:  world,
	}
}

func (n *NavigationMesh) Vertices() []mgl64.Vec3 {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	return n.vertices
}

func (n *NavigationMesh) Normals() []mgl64.Vec3 {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	return n.normals
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
	voxelDimension := 1.0
	delta := n.Volume.MaxVertex.Sub(n.Volume.MinVertex)
	var runs [3]int = [3]int{int(delta[0] / voxelDimension), int(delta[1] / voxelDimension), int(delta[2] / voxelDimension)}

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
					MinVertex: n.Volume.MinVertex.Add(mgl64.Vec3{float64(i), float64(j), float64(k)}.Mul(voxelDimension)),
					MaxVertex: n.Volume.MinVertex.Add(mgl64.Vec3{float64(i + 1), float64(j + 1), float64(k + 1)}.Mul(voxelDimension)),
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

	var voxelField [][][]Voxel
	voxelField = make([][][]Voxel, runs[0])
	for i := range voxelField {
		voxelField[i] = make([][]Voxel, runs[1])
		for j := range voxelField[i] {
			voxelField[i][j] = make([]Voxel, runs[2])
		}
	}

	go func() {
		voxelCount := 0
		for work := range outputWork {
			voxelCount++
			n.mutex.Lock()
			vertices, normals := genVertexRenderData(work.boundingBox)
			n.vertices = append(n.vertices, vertices...)
			n.normals = append(n.normals, normals...)
			n.mutex.Unlock()

			x, y, z := work.x, work.y, work.z
			voxelField[x][y][z] = Voxel{
				Filled: true,
				X:      x,
				Y:      y,
				Z:      z,
			}
		}
		fmt.Printf("generated %d voxels\n", voxelCount)
	}()

	for i := 0; i < runs[0]; i++ {
		for j := 0; j < runs[1]; j++ {
			for k := 0; k < runs[2]; k++ {
				inputWork <- [3]int{i, j, k}
			}
		}
	}
	close(inputWork)

	totalTricount := 0
	for _, count := range entityTriCount {
		totalTricount += count
	}
	fmt.Println("nav mesh entity tri count", totalTricount)
	return voxelField
}

func (n *NavigationMesh) buildNavigableArea(voxelField [][][]Voxel) [][][]Voxel {
	return voxelField
}

func (n *NavigationMesh) computeDistanceTransform(voxelField [][][]Voxel) [][][]int {
	return nil
}

func (n *NavigationMesh) BakeNavMesh() {
	go func() {
		voxelField := n.voxelize()
		voxelField = n.buildNavigableArea(voxelField)
		distanceTransform := n.computeDistanceTransform(voxelField)
		_ = distanceTransform
	}()
}

type Voxel struct {
	Filled  bool
	X, Y, Z int
}

type OutputWork struct {
	x, y, z     int
	boundingBox collider.BoundingBox
}
