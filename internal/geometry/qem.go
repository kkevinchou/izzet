package geometry

import (
	"fmt"
	"slices"
	"sort"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/collision/collider"
	"github.com/kkevinchou/izzet/internal/gheap"
	"github.com/kkevinchou/kitolib/modelspec"
	"github.com/kkevinchou/kitolib/utils"
	"github.com/patrick-higgins/rtreego"
)

// TODO - we should check that a merged edge doesn't cause the mesh to be degenerate. if so, skip
func SimplifyMesh(primitive *modelspec.PrimitiveSpecification, iterations int) *collider.TriMesh {
	allVertIndices, allVertPositions, v2t, triangles := mergeVerticesByDistance(primitive, 0.00001)

	// 1. Compute the Q matrices for all the initial vertices.
	quadrics := map[int]mgl64.Mat4{}
	for _, triangle := range sortedTriangles(triangles) {
		plane, ok := PlaneFromVerts([3]mgl64.Vec3{
			allVertPositions[triangle[0]],
			allVertPositions[triangle[1]],
			allVertPositions[triangle[2]],
		})

		if !ok {
			continue
		}

		q := ComputeErrorQuadric(plane)
		for _, vertex := range triangle {
			quadrics[vertex] = quadrics[vertex].Add(q)
		}
	}

	for _, i := range allVertIndices {
		qem := ComputeQEM(allVertPositions[i].Vec4(1), quadrics[int(i)])
		if qem > 1 {
			panic("qem of the vertex versus its own quadric should be 0")
		}
	}

	// 2. Select all valid pairs.
	// 3. Compute the optimal contraction target v¯ for each valid pair
	//	(v1, v2 ). The error v¯T(Q1 +Q2 )v¯ of this target vertex becomes the cost of contracting that pair.
	// 4. Place all the pairs in a heap keyed on cost with the minimum cost pair at the top.

	heap := gheap.New[*EdgeContraction](Less)

	validEdges := map[string]*EdgeContraction{}

	for _, triangle := range sortedTriangles(triangles) {
		for i := 0; i < 3; i++ {
			idx1 := triangle[i]
			idx2 := triangle[(i+1)%3]

			hash := minEdgeHash(idx1, idx2)
			if _, ok := validEdges[hash]; ok {
				continue
			}

			edgeContraction := createEdgeContraction(idx1, idx2, quadrics[int(idx1)], quadrics[int(idx2)], allVertPositions[idx1], allVertPositions[idx2])
			validEdges[hash] = edgeContraction
			heap.Push(edgeContraction)
		}
	}

	var bad bool
	var triMesh collider.TriMesh

	// 5. Iteratively remove the pair (v1, v2) of least cost from the heap, contract this pair, and update the costs of all valid pairs involving v1.
	for heap.Len() > 0 {
		edgeContraction := heap.Pop()
		if !edgeContraction.Valid {
			continue
		}

		// if math.Abs(edgeContraction.Cost) < 1 {
		// 	edgeContraction.Valid = false
		// 	continue
		// }

		if iterations <= 0 {
			break
		}

		iterations--

		var trisToDelete []int
		v1 := edgeContraction.Idx1
		v2 := edgeContraction.Idx2
		triMesh.DebugPoints = append(triMesh.DebugPoints, allVertPositions[v1], allVertPositions[v2])

		for _, triangleIdx := range v2t[v1] {
			for _, vert := range triangles[triangleIdx] {
				if vert == v1 {
					continue
				}
				if vert == v2 {
					trisToDelete = append(trisToDelete, triangleIdx)
				}
			}
		}

		if len(trisToDelete) == 1 {
			if bad {
				break
			}
			k := 1
			_ = k
			_ = bad
		}

		newVertexIndex := len(allVertPositions)
		allVertPositions = append(allVertPositions, edgeContraction.NewVertex)

		// mark edges incident to v1 and v2 as invalid

		v1Neighbors := vertNeighbors(v1, v2t, triangles)
		for _, neighbor := range v1Neighbors {
			hash := minEdgeHash(v1, neighbor)
			edgeContraction := validEdges[hash]
			edgeContraction.Valid = false
		}

		v2Neighbors := vertNeighbors(v2, v2t, triangles)
		for _, neighbor := range v2Neighbors {
			hash := minEdgeHash(v2, neighbor)
			edgeContraction := validEdges[hash]
			edgeContraction.Valid = false
		}

		for _, triToDelete := range trisToDelete {
			// delete the triangles
			for _, vert := range triangles[triToDelete] {
				v2t[vert] = slices.DeleteFunc(v2t[vert], func(t int) bool { return t == triToDelete })
			}
			delete(triangles, triToDelete)
		}

		// recalculate Q for all vNeighbors

		// update triangles that were incident to v1 v2 to now point to vHat instead
		replaceVertex(v1, newVertexIndex, v2t, triangles, validEdges)
		replaceVertex(v2, newVertexIndex, v2t, triangles, validEdges)

		// create Q for vHat
		q := createQuadricForVertex(newVertexIndex, allVertPositions, v2t, triangles)
		quadrics[newVertexIndex] = q

		// update Q for all neighbors
		neighbors := vertNeighbors(newVertexIndex, v2t, triangles)
		for _, neighbor := range neighbors {
			q := createQuadricForVertex(neighbor, allVertPositions, v2t, triangles)
			quadrics[neighbor] = q

			// push an edge contraction onto the heap for each vHat vNeighbor pair
			edgeContraction := createEdgeContraction(newVertexIndex, neighbor, quadrics[newVertexIndex], quadrics[neighbor], allVertPositions[newVertexIndex], allVertPositions[neighbor])
			if edgeContraction.NewVertex.X() > 30 {
				asdf := 1
				_ = asdf
			}

			hash := minEdgeHash(newVertexIndex, neighbor)
			validEdges[hash] = edgeContraction
			heap.Push(edgeContraction)
		}
	}

	for _, triangle := range sortedTriangles(triangles) {
		triMesh.Triangles = append(triMesh.Triangles,
			collider.NewTriangle(
				[3]mgl64.Vec3{
					allVertPositions[triangle[0]],
					allVertPositions[triangle[1]],
					allVertPositions[triangle[2]],
				},
			),
		)
	}

	return &triMesh
}

func sortedTriangles(triangles map[int][3]int) [][3]int {
	var indices []int
	for i := range triangles {
		indices = append(indices, i)
	}

	sort.Ints(indices)

	var result [][3]int

	for _, i := range indices {
		result = append(result, triangles[i])
	}

	return result
}

type Point struct {
	Index    int
	RPoint   rtreego.Point
	Distance float64
}

func (p Point) Bounds() *rtreego.Rect {
	return p.RPoint.ToRect(p.Distance)
}

func mergeVerticesByDistance(primitive *modelspec.PrimitiveSpecification, distance float64) ([]int, []mgl64.Vec3, map[int][]int, map[int][3]int) {
	var allVertPositions []mgl64.Vec3
	for _, v := range primitive.UniqueVertices {
		allVertPositions = append(allVertPositions, utils.Vec3F32ToF64(v.Position))
	}

	tree := rtreego.NewTree(1, 100)

	var merges int
	var allVertIndices []int
	for _, v := range primitive.VertexIndices {
		position := allVertPositions[int(v)]
		p := Point{Index: int(v), RPoint: rtreego.Point{position[0], position[1], position[2]}, Distance: distance}
		results := tree.SearchIntersect(p.Bounds())

		if len(results) > 0 {
			foundPoint := results[0].(Point)
			allVertIndices = append(allVertIndices, foundPoint.Index)
			merges += 1
		} else {
			tree.Insert(p)
			allVertIndices = append(allVertIndices, int(v))
		}
	}

	// geometry setup
	v2t := map[int][]int{}
	triangles := map[int][3]int{}
	triangleCounter := 0

	for i := 0; i < len(allVertIndices); i += 3 {
		v1 := allVertIndices[i]
		v2 := allVertIndices[i+1]
		v3 := allVertIndices[i+2]

		tri := [3]int{int(v1), int(v2), int(v3)}

		triangles[triangleCounter] = tri

		v2t[v1] = append(v2t[v1], triangleCounter)
		v2t[v2] = append(v2t[v2], triangleCounter)
		v2t[v3] = append(v2t[v3], triangleCounter)

		triangleCounter += 1
	}

	return allVertIndices, allVertPositions, v2t, triangles
}

func minEdgeHash(v1, v2 int) string {
	return fmt.Sprintf("%d_%d", min(v1, v2), max(v1, v2))
}

func createQuadricForVertex(vert int, allVertPositions []mgl64.Vec3, v2t map[int][]int, triangles map[int][3]int) mgl64.Mat4 {
	var q mgl64.Mat4
	for _, triIndex := range v2t[vert] {
		triangle := triangles[triIndex]
		plane, ok := PlaneFromVerts([3]mgl64.Vec3{
			allVertPositions[triangle[0]],
			allVertPositions[triangle[1]],
			allVertPositions[triangle[2]],
		})

		if !ok {
			continue
		}

		q = q.Add(ComputeErrorQuadric(plane))
	}
	return q
}

func replaceVertex(oldVert, newVert int, v2t map[int][]int, triangles map[int][3]int, validEdges map[string]*EdgeContraction) {
	for _, triIndex := range v2t[oldVert] {
		triCopy := triangles[triIndex]
		for i, val := range triCopy {
			if val == oldVert {
				triCopy[i] = newVert
				triangles[triIndex] = triCopy
				v2t[newVert] = append(v2t[newVert], triIndex)
				break
			}
		}
	}
}

func vertNeighbors(v1 int, v2t map[int][]int, triangles map[int][3]int) []int {
	var result []int
	seen := map[int]bool{}

	for _, tri := range v2t[v1] {
		for _, v2 := range triangles[tri] {
			if v2 != v1 {
				if _, ok := seen[v2]; ok {
					continue
				}
				seen[v2] = true
				result = append(result, v2)
			}
		}
	}

	return result
}

func createEdgeContraction(v1, v2 int, v1Quadric, v2Quadric mgl64.Mat4, v1Position, v2Position mgl64.Vec3) *EdgeContraction {
	idx1 := v1
	idx2 := v2

	QHat := v1Quadric.Add(v2Quadric)

	var vHat mgl64.Vec3
	var cost float64

	if QHat.Det() != 0 {
		vHatV4 := QHat.Inv().Mul4x1(mgl64.Vec4{0, 0, 0, 1})
		vHatV4 = vHatV4.Mul(1.0 / vHatV4.W())

		cost = ComputeQEM(vHatV4, QHat)
		vHat = vHatV4.Vec3()
	} else {
		// cost = 0
		// vHat = v1Position.Add(v2Position).Mul(1.0 / 2.0)
		// TODO - do the same check for the middle vertex
		v1Cost := ComputeQEM(v1Position.Vec4(1), QHat)
		v2Cost := ComputeQEM(v2Position.Vec4(1), QHat)

		if v1Cost < v2Cost {
			cost = v1Cost
			vHat = v1Position
		} else {
			cost = v2Cost
			vHat = v2Position
		}
	}

	return &EdgeContraction{Idx1: idx1, Idx2: idx2, NewVertex: vHat, Quadric: QHat, Cost: cost, Valid: true}
}

type EdgeContraction struct {
	Idx1      int
	Idx2      int
	NewVertex mgl64.Vec3
	Cost      float64
	Quadric   mgl64.Mat4
	Valid     bool
}

func Less(c1, c2 *EdgeContraction) bool {
	return c1.Cost < c2.Cost
}

func convertVertex(v modelspec.Vertex) mgl64.Vec4 {
	return utils.Vec3F32ToF64(v.Position).Vec4(1)
}
