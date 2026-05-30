package geometry_test

// func TestHalfEdgeSurfaceGeneration(t *testing.T) {
// 	config := &gltf.ParseConfig{TextureCoordStyle: gltf.TextureCoordStyleOpenGL}
// 	doc, err := gltf.ParseGLTF("model", "../../_assets/test/stall_manifold.gltf", config)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	var primitives []*modelspec.PrimitiveSpecification
// 	for _, m := range doc.Meshes {
// 		primitives = append(primitives, m.Primitives...)
// 	}
// 	surface := geometry.CreateHalfEdgeSurface(primitives)
// 	if len(surface.HalfEdges) == 0 {
// 		t.Error("surface has 0 half edges")
// 	}
// }
