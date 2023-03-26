package navmesh

import (
	"fmt"
	"strings"
	"sync"

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

	// navmesh     *pathing.NavMesh
	clusterIDs  []int
	constructed bool

	vertices []mgl64.Vec3

	voxels []collider.BoundingBox
	mutex  sync.Mutex
}

func New(world World) *NavigationMesh {
	return &NavigationMesh{
		Volume: collider.BoundingBox{MinVertex: mgl64.Vec3{-150, -150, -150}, MaxVertex: mgl64.Vec3{150, 150, 150}},
		world:  world,
	}
}

type RenderData struct {
}

func (n *NavigationMesh) RenderData() ([]mgl64.Vec3, []int) {
	return n.vertices, n.clusterIDs
}

func (n *NavigationMesh) Voxels() []collider.BoundingBox {
	return n.voxels
}
func (n *NavigationMesh) Vertices() []mgl64.Vec3 {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	return n.vertices
}

func (n *NavigationMesh) Voxelize() {
	spatialPartition := n.world.SpatialPartition()
	sEntities := spatialPartition.QueryEntities(n.Volume)

	if len(sEntities) == 0 {
		fmt.Println("not constructed")
		return
	}

	n.constructed = true
	var entities []*entities.Entity
	boundingBoxes := map[int]collider.BoundingBox{}
	meshTriangles := map[int][]Triangle{}
	entityTriCount := map[int]int{}

	for _, entity := range sEntities {
		e := n.world.GetEntityByID(entity.GetID())
		if !strings.Contains(e.Name, "Tile") {
			continue
		}

		// TEMP
		if e.GetID() != 6 {
			continue
		}

		entities = append(entities, e)
		boundingBoxes[e.GetID()] = *e.BoundingBox()

		for _, rd := range e.Model.RenderData() {
			meshID := rd.MeshID
			mesh := e.Model.Collection().Meshes[meshID]
			for i := 0; i < len(mesh.Vertices); i += 3 {
				meshTriangles[meshID] = append(meshTriangles[meshID],
					Triangle{
						V1: utils.Vec3F32ToF64(mesh.Vertices[i].Position),
						V2: utils.Vec3F32ToF64(mesh.Vertices[i+1].Position),
						V3: utils.Vec3F32ToF64(mesh.Vertices[i+2].Position),
					},
				)
			}
			numVerts := len(mesh.Vertices)
			entityTriCount[e.GetID()] = numVerts / 3
		}
	}

	outputWork := make(chan collider.BoundingBox)
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

				found := false
				for _, entity := range entities {
					bb := boundingBoxes[entity.GetID()]
					if !collision.CheckOverlapAABBAABB(voxel, &bb) {
						continue
					}

					for _, rd := range entity.Model.RenderData() {
						for _, tri := range meshTriangles[rd.MeshID] {
							if IntersectAABBTriangle(voxelAABB, tri) {
								outputWork <- *voxel
								// if a voxel is output, we don't need to check the rest of the entities
								found = true
							}
							if found {
								break
							}
						}
						if found {
							break
						}
					}
					if found {
						break
					}
				}
			}

			doneWorkerMutex.Lock()
			doneWorkerCount++
			if doneWorkerCount == workerCount {
				close(outputWork)
			}
			doneWorkerMutex.Unlock()
		}()
	}

	go func() {
		for voxel := range outputWork {
			n.voxels = append(n.voxels, voxel)
			n.mutex.Lock()
			n.vertices = append(n.vertices, bbVerts(voxel)...)
			n.mutex.Unlock()
		}
		fmt.Printf("generated %d voxels\n", len(n.voxels))
	}()

	// NOTE - with each entity's bounding box, i should be able to only create input work that overlaps with
	// the voxels i care about, rather than going through each voxel and doing bounding box checks against each entity
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
}

func (n *NavigationMesh) BakeNavMesh() {
	if n.constructed {
		return
	}
	n.Voxelize()
}
