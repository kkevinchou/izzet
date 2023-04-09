package navmesh

import (
	"testing"

	"github.com/go-gl/mathgl/mgl64"
)

func TestProjection(t *testing.T) {
	v1 := mgl64.Vec2{300.4, 100.4}
	v2 := mgl64.Vec2{0, 0}
	v3 := mgl64.Vec2{100, 0}
	FillBottomFlatTriangle(v1, v2, v3)
	t.Fail()
}
