package modellibrary

import (
	"sort"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/kkevinchou/izzet/izzet/model"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/kitolib/modelspec"
)

type Handle struct {
	namespace string
	id        int
}

func NewHandle(namespace string, id int) Handle {
	return Handle{namespace: namespace, id: id}
}

// Interface
// - stores models that have been loaded and builds their VAOs for later rendering
// - models can be referenced via string handle to fetch the associated model / vaos
// - this is used for rendering instantiated entities as well as serialization/deserialization

// Asset Manager
// - represents the model data from disk, agnostic of rendering backend (OpenGL, Vulkan, etc)

// Model Library
// - Multiple implementations, could be backed by OpenGL, Vulkan, etc

// Open Questions
// - is what's the source of truth for mesh data? the asset manager or the model library
// - hierarchical information related to parenting is good hinting that if we construct
// 		the parent entity, we should construct the child entities as well.
//			- e.g. if we construct the base of a house and the roof and walls should come
//				with it as well (since it's parented to the base)
//			- where should this hierarchical information be stored though?

// Notes
// - nodes
//		- have a name
// 		- references a single mesh
// - meshes
//		- have a name
// 		- references a collection of primitives
// - primitives
//		- have a name (not actually useful?)
// 		- VAOs are constructed at the primitive level
//		- each can have their own vertex attributes, materials, etc
// - primitives don't exist in a vacuumm and are always associated with a mesh
// - the model library should be a mesh library using mesh handles (their nane)

type Primitive struct {
	// Name      string
	Primitive *modelspec.PrimitiveSpecification
	Transform mgl32.Mat4

	// vao that contains all vertex attributes
	// position, normals, texture coords, joint indices/weights, etc
	VAO uint32

	// vao that only contains geometry related vertex attributes
	// i.e. vertex positions and joint indices / weights
	// but not normals, texture coords
	GeometryVAO uint32
}

type ModelLibrary struct {
	Primitives map[Handle][]Primitive
}

func New() *ModelLibrary {
	m := &ModelLibrary{
		Primitives: map[Handle][]Primitive{},
	}

	m.RegisterMesh("global", cube())

	return m
}

func (m *ModelLibrary) RegisterMesh(namespace string, mesh *modelspec.MeshSpecification) {
	modelConfig := &model.ModelConfig{MaxAnimationJointWeights: settings.MaxAnimationJointWeights}
	vaos := createVAOs(modelConfig, []*modelspec.MeshSpecification{mesh})
	geometryVAOs := createGeometryVAOs(modelConfig, []*modelspec.MeshSpecification{mesh})

	handle := NewHandle(namespace, mesh.ID)
	for i, primitive := range mesh.Primitives {
		m.Primitives[handle] = append(m.Primitives[handle], Primitive{
			Primitive:   primitive,
			VAO:         vaos[0][i],
			GeometryVAO: geometryVAOs[0][i],
		})
	}
}

func (m *ModelLibrary) GetPrimitives(handle Handle) []Primitive {
	return m.Primitives[handle]
}

// TODO: create the ability to create a singular vao that has all the mesh data merged into one
// also, when we merged everything into one vao, we need to first apply any node transformations
// onto the vertices since we won't be able to set uniforms for each mesh, as we now render them
// all at once, rather than one at a time and setting the transform uniforms
func createVAOs(modelConfig *model.ModelConfig, meshes []*modelspec.MeshSpecification) [][]uint32 {
	vaos := [][]uint32{}
	for i, mesh := range meshes {
		vaos = append(vaos, []uint32{})
		for _, p := range mesh.Primitives {
			// initialize the VAO
			var vao uint32
			gl.GenVertexArrays(1, &vao)
			gl.BindVertexArray(vao)
			vaos[i] = append(vaos[i], vao)

			var vertexAttributes []float32
			var jointIDsAttribute []int32
			var jointWeightsAttribute []float32

			// set up the source data for the VBOs
			for _, vertex := range p.UniqueVertices {
				position := vertex.Position
				normal := vertex.Normal
				texture0Coords := vertex.Texture0Coords
				texture1Coords := vertex.Texture1Coords
				jointIDs := vertex.JointIDs
				jointWeights := vertex.JointWeights

				vertexAttributes = append(vertexAttributes,
					position.X(), position.Y(), position.Z(),
					normal.X(), normal.Y(), normal.Z(),
					texture0Coords.X(), texture0Coords.Y(),
					texture1Coords.X(), texture1Coords.Y(),
				)

				ids, weights := fillWeights(jointIDs, jointWeights, modelConfig.MaxAnimationJointWeights)
				for _, id := range ids {
					jointIDsAttribute = append(jointIDsAttribute, int32(id))
				}
				jointWeightsAttribute = append(jointWeightsAttribute, weights...)
			}

			totalAttributeSize := len(vertexAttributes) / len(p.UniqueVertices)

			// lay out the position, normal, texture (index 0 and 1) coords in a VBO
			var vbo uint32
			gl.GenBuffers(1, &vbo)
			gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
			gl.BufferData(gl.ARRAY_BUFFER, len(vertexAttributes)*4, gl.Ptr(vertexAttributes), gl.STATIC_DRAW)

			ptrOffset := 0
			floatSize := 4

			// position
			gl.VertexAttribPointer(0, 3, gl.FLOAT, false, int32(totalAttributeSize)*4, nil)
			gl.EnableVertexAttribArray(0)

			ptrOffset += 3

			// normal
			gl.VertexAttribPointer(1, 3, gl.FLOAT, false, int32(totalAttributeSize)*4, gl.PtrOffset(ptrOffset*floatSize))
			gl.EnableVertexAttribArray(1)

			ptrOffset += 3

			// texture coords 0
			gl.VertexAttribPointer(2, 2, gl.FLOAT, false, int32(totalAttributeSize)*4, gl.PtrOffset(ptrOffset*floatSize))
			gl.EnableVertexAttribArray(2)

			ptrOffset += 2

			// texture coords 1
			gl.VertexAttribPointer(3, 2, gl.FLOAT, false, int32(totalAttributeSize)*4, gl.PtrOffset(ptrOffset*floatSize))
			gl.EnableVertexAttribArray(3)

			// lay out the joint IDs in a VBO
			var vboJointIDs uint32
			gl.GenBuffers(1, &vboJointIDs)
			gl.BindBuffer(gl.ARRAY_BUFFER, vboJointIDs)
			gl.BufferData(gl.ARRAY_BUFFER, len(jointIDsAttribute)*4, gl.Ptr(jointIDsAttribute), gl.STATIC_DRAW)
			gl.VertexAttribIPointer(4, int32(modelConfig.MaxAnimationJointWeights), gl.INT, int32(modelConfig.MaxAnimationJointWeights)*4, nil)
			gl.EnableVertexAttribArray(4)

			// lay out the joint weights in a VBO
			var vboJointWeights uint32
			gl.GenBuffers(1, &vboJointWeights)
			gl.BindBuffer(gl.ARRAY_BUFFER, vboJointWeights)
			gl.BufferData(gl.ARRAY_BUFFER, len(jointWeightsAttribute)*4, gl.Ptr(jointWeightsAttribute), gl.STATIC_DRAW)
			gl.VertexAttribPointer(5, int32(modelConfig.MaxAnimationJointWeights), gl.FLOAT, false, int32(modelConfig.MaxAnimationJointWeights)*4, nil)
			gl.EnableVertexAttribArray(5)

			// set up the EBO, each triplet of indices point to three vertices
			// that form a triangle.
			var ebo uint32
			gl.GenBuffers(1, &ebo)
			gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
			gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(p.VertexIndices)*4, gl.Ptr(p.VertexIndices), gl.STATIC_DRAW)
		}
	}

	return vaos
}

func createGeometryVAOs(modelConfig *model.ModelConfig, meshes []*modelspec.MeshSpecification) [][]uint32 {
	vaos := [][]uint32{}
	for i, mesh := range meshes {
		vaos = append(vaos, []uint32{})
		for _, p := range mesh.Primitives {
			// initialize the VAO
			var vao uint32
			gl.GenVertexArrays(1, &vao)
			gl.BindVertexArray(vao)
			vaos[i] = append(vaos[i], vao)

			var vertexAttributes []float32
			var jointIDsAttribute []int32
			var jointWeightsAttribute []float32

			// set up the source data for the VBOs
			for _, vertex := range p.UniqueVertices {
				position := vertex.Position
				jointIDs := vertex.JointIDs
				jointWeights := vertex.JointWeights

				vertexAttributes = append(vertexAttributes,
					position.X(), position.Y(), position.Z(),
				)

				ids, weights := fillWeights(jointIDs, jointWeights, modelConfig.MaxAnimationJointWeights)
				for _, id := range ids {
					jointIDsAttribute = append(jointIDsAttribute, int32(id))
				}
				jointWeightsAttribute = append(jointWeightsAttribute, weights...)
			}

			totalAttributeSize := len(vertexAttributes) / len(p.UniqueVertices)

			// lay out the position, normal, texture (index 0 and 1) coords in a VBO
			var vbo uint32
			gl.GenBuffers(1, &vbo)
			gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
			gl.BufferData(gl.ARRAY_BUFFER, len(vertexAttributes)*4, gl.Ptr(vertexAttributes), gl.STATIC_DRAW)

			ptrOffset := 0

			// position
			gl.VertexAttribPointer(0, 3, gl.FLOAT, false, int32(totalAttributeSize)*4, nil)
			gl.EnableVertexAttribArray(0)

			ptrOffset += 3

			// lay out the joint IDs in a VBO
			var vboJointIDs uint32
			gl.GenBuffers(1, &vboJointIDs)
			gl.BindBuffer(gl.ARRAY_BUFFER, vboJointIDs)
			gl.BufferData(gl.ARRAY_BUFFER, len(jointIDsAttribute)*4, gl.Ptr(jointIDsAttribute), gl.STATIC_DRAW)
			gl.VertexAttribIPointer(1, int32(modelConfig.MaxAnimationJointWeights), gl.INT, int32(modelConfig.MaxAnimationJointWeights)*4, nil)
			gl.EnableVertexAttribArray(1)

			// lay out the joint weights in a VBO
			var vboJointWeights uint32
			gl.GenBuffers(1, &vboJointWeights)
			gl.BindBuffer(gl.ARRAY_BUFFER, vboJointWeights)
			gl.BufferData(gl.ARRAY_BUFFER, len(jointWeightsAttribute)*4, gl.Ptr(jointWeightsAttribute), gl.STATIC_DRAW)
			gl.VertexAttribPointer(2, int32(modelConfig.MaxAnimationJointWeights), gl.FLOAT, false, int32(modelConfig.MaxAnimationJointWeights)*4, nil)
			gl.EnableVertexAttribArray(2)

			// set up the EBO, each triplet of indices point to three vertices
			// that form a triangle.
			var ebo uint32
			gl.GenBuffers(1, &ebo)
			gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
			gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(p.VertexIndices)*4, gl.Ptr(p.VertexIndices), gl.STATIC_DRAW)
		}
	}

	return vaos
}

func fillWeights(jointIDs []int, weights []float32, maxAnimationJointWeights int) ([]int, []float32) {
	j := []int{}
	w := []float32{}

	if len(jointIDs) <= maxAnimationJointWeights {
		j = append(j, jointIDs...)
		w = append(w, weights...)
		// fill in empty jointIDs and weights
		for i := 0; i < maxAnimationJointWeights-len(jointIDs); i++ {
			j = append(j, 0)
			w = append(w, 0)
		}
	} else if len(jointIDs) > maxAnimationJointWeights {
		jointWeights := []JointWeight{}
		for i := range jointIDs {
			jointWeights = append(jointWeights, JointWeight{JointID: jointIDs[i], Weight: weights[i]})
		}
		sort.Sort(sort.Reverse(byWeights(jointWeights)))

		// take top 3 weights
		jointWeights = jointWeights[:maxAnimationJointWeights]
		NormalizeWeights(jointWeights)
		for _, jw := range jointWeights {
			j = append(j, jw.JointID)
			w = append(w, jw.Weight)
		}
	}

	return j, w
}

func NormalizeWeights(jointWeights []JointWeight) {
	var totalWeight float32
	for _, jw := range jointWeights {
		totalWeight += jw.Weight
	}

	for i := range jointWeights {
		jointWeights[i].Weight /= totalWeight
	}
}

type byWeights []JointWeight

type JointWeight struct {
	JointID int
	Weight  float32
}

func (s byWeights) Len() int {
	return len(s)
}
func (s byWeights) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s byWeights) Less(i, j int) bool {
	return s[i].Weight < s[j].Weight
}
