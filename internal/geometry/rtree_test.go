package geometry_test

import (
	"fmt"
	"testing"

	"github.com/kkevinchou/izzet/internal/geometry"
	"github.com/kkevinchou/kitolib/assets/loaders/gltf"
	"github.com/patrick-higgins/rtreego"
)

type Point struct {
	rtreego.Point
}

func (p Point) Bounds() *rtreego.Rect {
	return p.Point.ToRect(0.1)
}

func TestRTree(t *testing.T) {
	rt := rtreego.NewTree(2, 25)

	p1 := Point{Point: rtreego.Point{0, 0, 0}}
	p2 := Point{Point: rtreego.Point{1, 0, 0}}
	p3 := Point{Point: rtreego.Point{2, 0, 0}}

	rt.Insert(p1)
	rt.Insert(p2)
	rt.Insert(p3)

	results := rt.SearchIntersect(rtreego.Point{0, 0, 0}.ToRect(1))
	fmt.Println(results)
}

func TestMergeStall(t *testing.T) {
	config := &gltf.ParseConfig{TextureCoordStyle: gltf.TextureCoordStyleOpenGL}
	doc, err := gltf.ParseGLTF("model", "../../_assets/test/stall.gltf", config)
	if err != nil {
		t.Fail()
		t.Errorf(err.Error())
	}

	geometry.SimplifyMesh(doc.Meshes[0].Primitives[0], -1)
}
