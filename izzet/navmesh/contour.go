package navmesh

import "github.com/go-gl/mathgl/mgl64"

type Contour struct {
	//
	//
}

type ContourSet struct {
	bMin, bMax    mgl64.Vec3
	NumContours   int
	width, height int
	maxError      float64
	contours      []Contour
}

func BuildContours(chf *CompactHeightField, maxError float64, maxEdgeLength int) *ContourSet {
	contourSet := &ContourSet{
		bMin:     chf.bMin,
		bMax:     chf.bMax,
		width:    chf.width,
		height:   chf.height,
		maxError: maxError,
		contours: make([]Contour, chf.maxRegions),
	}

	return contourSet
}
