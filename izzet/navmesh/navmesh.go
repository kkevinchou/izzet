package navmesh

import (
	"fmt"
	"math"
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

	constructed bool

	vertices []mgl64.Vec3
	normals  []mgl64.Vec3

	mutex sync.Mutex
}

func New(world World) *NavigationMesh {
	return &NavigationMesh{
		Volume: collider.BoundingBox{MinVertex: mgl64.Vec3{-150, -150, -150}, MaxVertex: mgl64.Vec3{150, 150, 150}},
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

func (n *NavigationMesh) Voxelize() {
	start := time.Now()
	spatialPartition := n.world.SpatialPartition()
	sEntities := spatialPartition.QueryEntities(n.Volume)

	if len(sEntities) == 0 {
		fmt.Println("not constructed")
		return
	}

	n.constructed = true
	var candidateEntities []*entities.Entity
	boundingBoxes := map[int]collider.BoundingBox{}
	meshTriangles := map[int][]Triangle{}
	entityTriCount := map[int]int{}

	for _, entity := range sEntities {
		e := n.world.GetEntityByID(entity.GetID())

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
					MinVertex: mgl64.Vec3{float64(i), float64(j), float64(k)},
					MaxVertex: mgl64.Vec3{float64(i) + voxelDimension, float64(j) + voxelDimension, float64(k) + voxelDimension},
				}

				voxelAABB := AABB{Min: voxel.MinVertex, Max: voxel.MaxVertex}
				found := false

				for _, entity := range candidateEntities {
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
				fmt.Println("finished work in", time.Since(start).Seconds())
				close(outputWork)
			}
			doneWorkerMutex.Unlock()
		}()
	}

	go func() {
		voxelCount := 0
		for voxel := range outputWork {
			voxelCount++
			n.mutex.Lock()
			vertices, normals := genVertexRenderData(voxel)
			n.vertices = append(n.vertices, vertices...)
			n.normals = append(n.normals, normals...)
			n.mutex.Unlock()
		}
		fmt.Printf("generated %d voxels\n", voxelCount)
	}()

	seen := map[[3]int]bool{}
	aabb2 := n.Volume
	for _, entity := range candidateEntities {
		aabb1 := boundingBoxes[entity.GetID()]

		if aabb1.MaxVertex.X() < aabb2.MinVertex.X() || aabb1.MinVertex.X() > aabb2.MaxVertex.X() {
			continue
		}

		if aabb1.MaxVertex.Y() < aabb2.MinVertex.Y() || aabb1.MinVertex.Y() > aabb2.MaxVertex.Y() {
			continue
		}

		if aabb1.MaxVertex.Z() < aabb2.MinVertex.Z() || aabb1.MinVertex.Z() > aabb2.MaxVertex.Z() {
			continue
		}

		minX := int(math.Max(aabb1.MinVertex.X(), aabb2.MinVertex.X()) - voxelDimension)
		maxX := int(math.Min(aabb1.MaxVertex.X(), aabb2.MaxVertex.X()) + voxelDimension)

		minY := int(math.Max(aabb1.MinVertex.Y(), aabb2.MinVertex.Y()) - voxelDimension)
		maxY := int(math.Min(aabb1.MaxVertex.Y(), aabb2.MaxVertex.Y()) + voxelDimension)

		minZ := int(math.Max(aabb1.MinVertex.Z(), aabb2.MinVertex.Z()) - voxelDimension)
		maxZ := int(math.Min(aabb1.MaxVertex.Z(), aabb2.MaxVertex.Z()) + voxelDimension)

		for i := minX; i < maxX; i++ {
			for j := minY; j < maxY; j++ {
				for k := minZ; k < maxZ; k++ {
					if _, ok := seen[[3]int{i, j, k}]; ok {
						continue
					}

					seen[[3]int{i, j, k}] = true
					inputWork <- [3]int{i, j, k}
				}
			}
		}

		// x := minX
		// y := minY
		// z := minZ

		// for x < maxX+voxelDimension {
		// 	y = minY
		// 	for y < maxY+voxelDimension {
		// 		z = minZ
		// 		for z < maxZ+voxelDimension {
		// 			if _, ok := seen[[3]int{int(x), int(y), int(z)}]; !ok {
		// 				inputWork <- [3]float64{x, y, z}
		// 			}
		// 			z += voxelDimension
		// 		}
		// 		y += voxelDimension
		// 	}
		// 	x += voxelDimension
		// }

		// if !collision.CheckOverlapAABBAABB(voxel, &bb) {
		// 	continue
		// }
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
