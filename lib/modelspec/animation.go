package modelspec

import (
	"time"

	"github.com/go-gl/mathgl/mgl32"
)

// This is static data that we read from the animation data files
type AnimationSpec struct {
	Name      string
	KeyFrames []*KeyFrame
	Length    time.Duration
}

// KeyFrame contains a "Pose" which is the mapping from joint index to
// the transformations that should be applied to the joint for this pose
type KeyFrame struct {
	Pose  map[int]*JointTransform
	Start time.Duration
}

// JointTransform represents the joint-space transformations that should be
// applied to the joint for the KeyFrame it is associated with.
type JointTransform struct {
	Translation mgl32.Vec3
	Rotation    mgl32.Quat
	Scale       mgl32.Vec3
}

func NewDefaultJointTransform() *JointTransform {
	return &JointTransform{
		Translation: mgl32.Vec3{0, 0, 0},
		Rotation:    mgl32.QuatIdent(),
		Scale:       mgl32.Vec3{1, 1, 1},
	}
}
