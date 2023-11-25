package geometry

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/kkevinchou/kitolib/modelspec"
)

type HalfEdgeSurface struct {
	HalfEdges []*HalfEdge
	Vertices  []Vertex
	// Edges    []Edge
	// Faces    []Face
}

type HalfEdge struct {
	Next   *HalfEdge
	Twin   *HalfEdge
	Vertex int
	// Edge   Edge
	// Face   Face
}

type Vertex struct {
	Position mgl32.Vec3
	HalfEdge *HalfEdge
}

// type Face struct {
// 	HalfEdge *HalfEdge
// }

// type Edge struct {
// 	HalfEdge *HalfEdge
// }

// collect all vertices
// compute Q for each tri (which defines a plane)
//	this can be reused for 3 other vertices
// collect all Qi from each tri incident to a vertex
// sum all Qi as Q and assign to the vertex
// compute the optimal contraction between v1 and v2 as vHat
//	vHat quadric is (Q1 + Q2)
// put all pairs in a heap
// iteratively remove pairs.
//	contract the pair
// 	update the error cost for all vertices connected to vHat
// stop after some number of contractions

func CreateHalfEdgeSurface(primitives []*modelspec.PrimitiveSpecification) *HalfEdgeSurface {
	for _, p := range primitives {
		numVertices := len(p.UniqueVertices)
		// numFaces := len(p.VertexIndices) / 3
		// numEdges := (len(p.VertexIndices) + numFaces - 2) / 2
		numEdges := 1000000
		numHalfEdges := numEdges * 2

		halfEdgeCache := map[string][2]*HalfEdge{}

		surface := &HalfEdgeSurface{
			Vertices: make([]Vertex, numVertices),
			// Edges:     make([]Edge, numEdges),
			// Faces:     make([]Face, numFaces),
			HalfEdges: make([]*HalfEdge, numHalfEdges),
		}

		for i, v := range p.UniqueVertices {
			surface.Vertices[i].Position = v.Position
		}

		// verts := p.UniqueVertices
		vertIndices := p.VertexIndices

		for i := 0; i < len(vertIndices); i += 3 {
			h1 := createOrLookupHalfEdges(int(vertIndices[i]), int(vertIndices[i+1]), surface, halfEdgeCache)
			h2 := createOrLookupHalfEdges(int(vertIndices[i+1]), int(vertIndices[i+2]), surface, halfEdgeCache)
			h3 := createOrLookupHalfEdges(int(vertIndices[i+2]), int(vertIndices[i]), surface, halfEdgeCache)

			h1.Next = h2
			h2.Next = h3
			h3.Next = h1
		}
		return surface
	}
	return nil
}

// // returns a face, half edges
// func process(vertices []Vertex, i, j, k int) (Edge, Face) {
// 	return Edge{}, Face{}
// }

var halfEdgeCursor int

func createOrLookupHalfEdges(v1, v2 int, surface *HalfEdgeSurface, halfEdgeCache map[string][2]*HalfEdge) *HalfEdge {
	// if v1 == 11409 || v1 == 11430 {
	// if v1 == 11409 || v2 == 11409 {
	// 	fmt.Println(v1, v2)
	// }

	k := edgeHash(v1, v2)
	if _, ok := halfEdgeCache[k]; ok {
		fmt.Println(k, "found twice")
	}

	twinKey := edgeHash(v2, v1)
	var mainHalfEdge *HalfEdge
	if _, ok := halfEdgeCache[twinKey]; !ok {
		halfEdge, twin := createHalfEdges(v1, v2, surface)
		halfEdgeArray := [2]*HalfEdge{
			halfEdge,
			twin,
		}

		currentHashKey := edgeHash(v1, v2)
		halfEdgeCache[currentHashKey] = halfEdgeArray
		mainHalfEdge = halfEdge
	} else {
		// if the twin key was found, then this half edge has already been created, return it
		mainHalfEdge = halfEdgeCache[twinKey][1]
	}
	return mainHalfEdge
}

func createHalfEdges(v1, v2 int, surface *HalfEdgeSurface) (*HalfEdge, *HalfEdge) {
	h1 := &HalfEdge{
		Vertex: v1,
	}
	h2 := &HalfEdge{
		Vertex: v2,
	}

	h1.Twin = h2
	h2.Twin = h2

	surface.HalfEdges[halfEdgeCursor] = h1
	surface.HalfEdges[halfEdgeCursor+1] = h2
	halfEdgeCursor += 2

	return h1, h2
}

func edgeHash(v1, v2 int) string {
	if v1 == v2 {
		panic("edge hash of the a vertex to itself")
	}
	return fmt.Sprintf("%d_%d", v1, v2)
}
