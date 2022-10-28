package libutils_test

import (
	"testing"

	"github.com/go-gl/mathgl/mgl64"
)

func TestQuat(t *testing.T) {
	v1 := mgl64.Vec3{0, 0, -1}
	v2 := mgl64.Vec3{0, 1, 0}

	mgl64.QuatBetweenVectors(v1, v2)
}
