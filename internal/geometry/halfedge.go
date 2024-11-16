package geometry

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/kkevinchou/kitolib/modelspec"
)

type HalfEdgeSurface struct {
	HalfEdges       []*HalfEdge
	Vertices        []Vertex
	VertsToHalfEdge map[int][]*HalfEdge
	NumHalfEdges    int
	// Edges    []Edge
	// Faces    []Face
}

type HalfEdge struct {
	Next       *HalfEdge
	Twin       *HalfEdge
	Vertex     int
	NextVertex int
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

// INITIALIZING

// sum all Qi (producing Q) and assign to the vertex
// create pairs of vertices if they meet at least one of the conditions:
//     1. each vertex in the pair are incident to each other
//     2. the vertices in the pair are within some distance threshold to another
// compute Q = Q1 + Q2
// find the optimal contraction between v1 and v2 (producing vhat)
//     if the determinant of Q is undefined, choose the best vertex between v1, v2, ()v1 + v2 / 2)
// place the error score and vertex pair in the contraction heap

// PROCESSING - processing the contraction heap

// pop from the contraction heap
// contract the pair
// update the position of v1 to vhat
// update the error cost for all vertices connected to vhat
//
// stop after we've reached a target number of triangles/ stop after some number of contractions

func CreateHalfEdgeSurface(primitives []*modelspec.PrimitiveSpecification) *HalfEdgeSurface {
	var halfEdgeCursor int
	p := primitives[0]

	numVertices := len(p.UniqueVertices)
	numEdges := len(p.Vertices)
	upperboundNumHalfEdges := numEdges * 2 // upperbound on the number of half edges, if you have two edges A -> B and B -> A then that'd only generate two half edges

	halfEdgeCache := map[string][2]*HalfEdge{}

	surface := &HalfEdgeSurface{
		Vertices:        make([]Vertex, numVertices),
		HalfEdges:       make([]*HalfEdge, upperboundNumHalfEdges),
		VertsToHalfEdge: make(map[int][]*HalfEdge),
	}

	for i, v := range p.UniqueVertices {
		surface.Vertices[i].Position = v.Position
	}

	vertIndices := p.VertexIndices

	for i := 0; i < len(vertIndices); i += 3 {
		h1 := createOrLookupHalfEdges(int(vertIndices[i]), int(vertIndices[i+1]), surface, halfEdgeCache, &halfEdgeCursor)
		h2 := createOrLookupHalfEdges(int(vertIndices[i+1]), int(vertIndices[i+2]), surface, halfEdgeCache, &halfEdgeCursor)
		h3 := createOrLookupHalfEdges(int(vertIndices[i+2]), int(vertIndices[i]), surface, halfEdgeCache, &halfEdgeCursor)

		h1.Next = h2
		h2.Next = h3
		h3.Next = h1

		surface.VertsToHalfEdge[h1.Vertex] = append(surface.VertsToHalfEdge[h1.Vertex], h1)
		surface.VertsToHalfEdge[h2.Vertex] = append(surface.VertsToHalfEdge[h2.Vertex], h2)
		surface.VertsToHalfEdge[h3.Vertex] = append(surface.VertsToHalfEdge[h3.Vertex], h3)
	}
	surface.NumHalfEdges = halfEdgeCursor

	createBoundaryHalfEdge(surface, p, halfEdgeCursor)

	return surface
}

func createBoundaryHalfEdge(surface *HalfEdgeSurface, primitive *modelspec.PrimitiveSpecification, halfEdgeCount int) {
	halfEdgesWithNilNext := map[int][]*HalfEdge{}
	for i := 0; i < halfEdgeCount; i++ {
		he := surface.HalfEdges[i]
		if he.Next != nil {
			continue
		}

		halfEdgesWithNilNext[he.Vertex] = append(halfEdgesWithNilNext[he.Vertex], he)
	}

	for _, halfEdges := range halfEdgesWithNilNext {
		for _, he := range halfEdges {
			if len(halfEdgesWithNilNext[he.NextVertex]) > 1 {
			} else {
				he.Next = halfEdgesWithNilNext[he.NextVertex][0]
			}
		}
	}

	count := 0
	for i := 0; i < halfEdgeCount; i++ {
		he := surface.HalfEdges[i]
		if he.Next != nil {
			continue
		}
		count += 1
	}

	// sort half edges by start
	// binary search to find the next half edge to connect
}

func createOrLookupHalfEdges(v1, v2 int, surface *HalfEdgeSurface, halfEdgeCache map[string][2]*HalfEdge, halfEdgeCursor *int) *HalfEdge {
	k := edgeHash(v1, v2)
	if _, ok := halfEdgeCache[k]; ok {
		panic("found twice")
	}

	twinKey := edgeHash(v2, v1)
	var mainHalfEdge *HalfEdge
	if _, ok := halfEdgeCache[twinKey]; !ok {
		halfEdge, twin := createHalfEdges(v1, v2, surface, halfEdgeCursor)
		halfEdgeArray := [2]*HalfEdge{
			halfEdge,
			twin,
		}

		halfEdgeCache[k] = halfEdgeArray
		mainHalfEdge = halfEdge
	} else {
		// if the twin key was found, then this half edge has already been created, return it
		mainHalfEdge = halfEdgeCache[twinKey][1]
	}
	return mainHalfEdge
}

func createHalfEdges(v1, v2 int, surface *HalfEdgeSurface, halfEdgeCursor *int) (*HalfEdge, *HalfEdge) {
	h1 := &HalfEdge{
		Vertex:     v1,
		NextVertex: v2,
	}
	h2 := &HalfEdge{
		Vertex:     v2,
		NextVertex: v1,
	}

	h1.Twin = h2
	h2.Twin = h2

	surface.HalfEdges[*halfEdgeCursor] = h1
	surface.HalfEdges[*halfEdgeCursor+1] = h2
	*halfEdgeCursor += 2

	return h1, h2
}

func edgeHash(v1, v2 int) string {
	if v1 == v2 {
		panic("edge hash of the a vertex to itself")
	}
	return fmt.Sprintf("%d_%d", v1, v2)
}
