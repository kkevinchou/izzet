package geometry_test

import (
	"fmt"
	"testing"

	"github.com/kkevinchou/izzet/lib/geometry"
	"github.com/kkevinchou/kitolib/assets/loaders/gltf"
	"github.com/kkevinchou/kitolib/modelspec"
)

func TestHalfEdgeSurfaceGeneration(t *testing.T) {
	config := &gltf.ParseConfig{TextureCoordStyle: gltf.TextureCoordStyleOpenGL}
	doc, err := gltf.ParseGLTF("model", "../../_assets/test/dude.gltf", config)
	// doc, err := gltf.ParseGLTF("model", "../../_assets/gltf/alpha.gltf", config)
	if err != nil {
		t.Fail()
		t.Errorf(err.Error())
	}

	var primitives []*modelspec.PrimitiveSpecification
	for _, m := range doc.Meshes {
		primitives = append(primitives, m.Primitives...)
	}
	surface := geometry.CreateHalfEdgeSurface(primitives)
	fmt.Println(surface)
}
