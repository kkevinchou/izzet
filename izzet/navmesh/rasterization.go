package navmesh

import (
	"fmt"
	"sort"
	"sync"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/kitolib/collision/collider"
	"github.com/kkevinchou/kitolib/utils"
)

type Span struct {
	Min   int
	Max   int
	Valid bool
}

func (n *NavigationMesh) voxelize() [][][]Voxel {
	// start := time.Now()
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

		// if !strings.Contains(e.Model.Name(), "Tile") && !strings.Contains(e.Model.Name(), "Stair") {
		// 	continue
		// }

		candidateEntities = append(candidateEntities, e)
		boundingBoxes[e.GetID()] = *e.BoundingBox()
		transform := utils.Mat4F64ToF32(entities.WorldTransform(e))

		for _, rd := range e.Model.RenderData() {
			meshID := rd.MeshID
			mesh := e.Model.Collection().Meshes[meshID]
			for i := 0; i < len(mesh.Vertices); i += 3 {
				worldSpaceV1 := utils.Vec3F32ToF64(transform.Mul4x1(mesh.Vertices[i].Position.Vec4(1)).Vec3())
				worldSpaceV2 := utils.Vec3F32ToF64(transform.Mul4x1(mesh.Vertices[i+1].Position.Vec4(1)).Vec3())
				worldSpaceV3 := utils.Vec3F32ToF64(transform.Mul4x1(mesh.Vertices[i+2].Position.Vec4(1)).Vec3())

				t := Triangle{
					V1: convertPointToVoxelFieldPosition(worldSpaceV1, n.Volume, n.voxelDimension),
					V2: convertPointToVoxelFieldPosition(worldSpaceV2, n.Volume, n.voxelDimension),
					V3: convertPointToVoxelFieldPosition(worldSpaceV3, n.Volume, n.voxelDimension),
				}
				meshTriangles[meshID] = append(meshTriangles[meshID], t)
			}
			numVerts := len(mesh.Vertices)
			entityTriCount[e.GetID()] = numVerts / 3
		}
	}

	delta := n.Volume.MaxVertex.Sub(n.Volume.MinVertex)
	var dimensions [3]int = [3]int{int(delta[0] / n.voxelDimension), int(delta[1] / n.voxelDimension), int(delta[2] / n.voxelDimension)}

	// initialize the voxel field
	voxelField := make([][][]Voxel, dimensions[0])
	for i := range voxelField {
		voxelField[i] = make([][]Voxel, dimensions[1])
		for j := range voxelField[i] {
			voxelField[i][j] = make([]Voxel, dimensions[2])
		}
	}

	tri := Triangle{
		A1: mgl64.Vec3{0, 5, 0},
		A2: mgl64.Vec3{25, 35, -25},
		A3: mgl64.Vec3{-5, 45, 25},
	}
	tri.V1 = convertPointToVoxelFieldPosition(tri.A1, n.Volume, n.voxelDimension)
	tri.V2 = convertPointToVoxelFieldPosition(tri.A2, n.Volume, n.voxelDimension)
	tri.V3 = convertPointToVoxelFieldPosition(tri.A3, n.Volume, n.voxelDimension)

	xSortedVerts := []mgl64.Vec3{tri.A1, tri.A2, tri.A3}
	sort.Slice(xSortedVerts, func(i, j int) bool {
		return xSortedVerts[i].X() < xSortedVerts[j].X()
	})
	tri.XSortedVerts = xSortedVerts

	zSortedVerts := []mgl64.Vec3{tri.A1, tri.A2, tri.A3}
	sort.Slice(zSortedVerts, func(i, j int) bool {
		return zSortedVerts[i].Z() < zSortedVerts[j].Z()
	})
	tri.ZSortedVerts = zSortedVerts

	ySortedVerts := []mgl64.Vec3{tri.A1, tri.A2, tri.A3}
	sort.Slice(ySortedVerts, func(i, j int) bool {
		return ySortedVerts[i].Y() < ySortedVerts[j].Y()
	})
	tri.YSortedVerts = ySortedVerts

	meshTriangles[69] = append(meshTriangles[69], tri)

	// leftRay := xSortedVerts[0].Sub(xSortedVerts[1]).Normalize()
	// rightRay := xSortedVerts[2].Sub(xSortedVerts[1]).Normalize()

	ray0 := zSortedVerts[2].Sub(zSortedVerts[0]).Normalize()
	ray1 := zSortedVerts[2].Sub(zSortedVerts[1]).Normalize()
	var zFloat float64 = zSortedVerts[2].Z()

	yzField := [100][100]Span{}

	vertRef0 := zSortedVerts[0]
	vertRef1 := zSortedVerts[1]

	for z := int(zSortedVerts[2].Z()); z >= int(zSortedVerts[0].Z()); z-- {
		if int(zSortedVerts[1].Z()) == int(zFloat) {
			ray1 = zSortedVerts[1].Sub(zSortedVerts[0]).Normalize()
			vertRef1 = vertRef0
		}
		delta0 := zFloat - vertRef0.Z()
		clippedVertex0 := vertRef0.Add(ray0.Mul(delta0 / ray0.Z()))
		clippedVoxel0 := convertPointToVoxelFieldPosition(clippedVertex0, n.Volume, n.voxelDimension)

		delta1 := zFloat - vertRef1.Z()
		clippedVertex1 := vertRef1.Add(ray1.Mul(delta1 / ray1.Z()))
		clippedVoxel1 := convertPointToVoxelFieldPosition(clippedVertex1, n.Volume, n.voxelDimension)

		if clippedVoxel0[0] < clippedVoxel1[0] {
			yzField[clippedVoxel0[1]][int(zFloat-n.Volume.MinVertex.Z())].Valid = true
			yzField[clippedVoxel0[1]][int(zFloat-n.Volume.MinVertex.Z())].Min = clippedVoxel0[0]

			yzField[clippedVoxel1[1]][int(zFloat-n.Volume.MinVertex.Z())].Valid = true
			yzField[clippedVoxel1[1]][int(zFloat-n.Volume.MinVertex.Z())].Max = clippedVoxel1[0]
		} else {
			yzField[clippedVoxel1[1]][int(zFloat-n.Volume.MinVertex.Z())].Valid = true
			yzField[clippedVoxel1[1]][int(zFloat-n.Volume.MinVertex.Z())].Min = clippedVoxel1[0]

			yzField[clippedVoxel0[1]][int(zFloat-n.Volume.MinVertex.Z())].Valid = true
			yzField[clippedVoxel0[1]][int(zFloat-n.Volume.MinVertex.Z())].Max = clippedVoxel0[0]
		}

		zFloat -= 1
	}

	for i := len(yzField) - 1; i >= 0; i-- {
		row := yzField[i]
		for _, val := range row {
			if val.Valid {
				fmt.Printf("x")
			} else {
				fmt.Printf("-")
			}
		}
		fmt.Printf("\n")
	}

	for id, triangles := range meshTriangles {
		if id != 69 {
			continue
		}
		for _, triangle := range triangles {
			// minXVertex, maxXVertex := triangle.V1, triangle.V1

			n.voxelCount += RasterizeLine(triangle.V1, triangle.V2, voxelField)
			n.voxelCount += RasterizeLine(triangle.V2, triangle.V3, voxelField)
			n.voxelCount += RasterizeLine(triangle.V3, triangle.V1, voxelField)
		}
	}
	n.voxelCount = 69

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

func RasterizeLine(start, end [3]int, voxelGrid [][][]Voxel) int {
	var voxelCount int
	direction := [3]int{end[0] - start[0], end[1] - start[1], end[2] - start[2]}
	dx, dy, dz := direction[0], direction[1], direction[2]

	// Determine the signs of dx, dy, and dz
	sx, sy, sz := sign(dx), sign(dy), sign(dz)

	// Determine which of the three dimensions has the greatest absolute difference
	adx, ady, adz := abs(dx), abs(dy), abs(dz)

	if adx >= ady && adx >= adz {
		// The X dimension is the major axis
		yd := ady - adx/2
		zd := adz - adx/2

		y, z := start[1], start[2]
		for x := start[0]; x != end[0]; x += sx {
			voxelCount += setVoxel(x, y, z, voxelGrid)

			yd += ady
			if yd >= adx {
				y += sy
				yd -= adx
			}

			zd += adz
			if zd >= adx {
				z += sz
				zd -= adx
			}
		}

	} else if ady >= adx && ady >= adz {
		// The Y dimension is the major axis
		xd := adx - ady/2
		zd := adz - ady/2

		x, z := start[0], start[2]
		for y := start[1]; y != end[1]; y += sy {
			voxelCount += setVoxel(x, y, z, voxelGrid)

			xd += adx
			if xd >= ady {
				x += sx
				xd -= ady
			}

			zd += adz
			if zd >= ady {
				z += sz
				zd -= ady
			}
		}

	} else {
		// The Z dimension is the major axis
		xd := adx - adz/2
		yd := ady - adz/2

		x, y := start[0], start[1]
		for z := start[2]; z != end[2]; z += sz {
			voxelCount += setVoxel(x, y, z, voxelGrid)

			xd += adx
			if xd >= adz {
				x += sx
				xd -= adz
			}

			yd += ady
			if yd >= adz {
				y += sy
				yd -= adz
			}
		}
	}
	return voxelCount
}

func setVoxel(x, y, z int, voxelGrid [][][]Voxel) int {
	voxelGrid[x][y][z] = NewVoxel(x, y, z)
	voxelGrid[x][y][z].Filled = true
	voxelGrid[x][y][z].RegionID = 69
	return 1
}

func sign(x int) int {
	if x > 0 {
		return 1
	} else if x < 0 {
		return -1
	} else {
		return 0
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	} else {
		return x
	}
}

func convertPointToVoxelFieldPosition(point mgl64.Vec3, volume collider.BoundingBox, voxelDimension float64) VoxelPosition {
	x := point.X() / voxelDimension
	y := point.Y() / voxelDimension
	z := point.Z() / voxelDimension

	return VoxelPosition{int(x - volume.MinVertex.X()), int(y - volume.MinVertex.Y()), int(z - volume.MinVertex.Z())}
}

// type MinXTriangles []Triangle

// func (u MinXTriangles) Len() int {
// 	return len(u)
// }
// func (u MinXTriangles) Swap(i, j int) {
// 	u[i], u[j] = u[j], u[i]
// }
// func (u MinXTriangles) Less(i, j int) bool {
// 	return u[i].MinX < u[j].MinX
// }
