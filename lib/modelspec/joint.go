package modelspec

import "github.com/go-gl/mathgl/mgl32"

type JointSpec struct {
	ID                   int
	Name                 string
	BindTransform        mgl32.Mat4
	InverseBindTransform mgl32.Mat4
	Children             []*JointSpec
}
