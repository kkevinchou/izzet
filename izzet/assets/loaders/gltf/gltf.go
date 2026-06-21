package gltf

import (
	"fmt"
	"path"
	"strings"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/kkevinchou/izzet/internal/modelspec"
	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"
)

func ParseGLTF(name string, documentPath string, config *ParseConfig) (*modelspec.Document, error) {
	var document modelspec.Document

	document.Name = name

	gltfDocument, err := gltf.Open(documentPath)
	if err != nil {
		return nil, err
	}
	ctx := newParseContext()

	document.PeripheralFiles, err = getPeripheralFiles(gltfDocument)
	if err != nil {
		return nil, err
	}

	var parsedJoints *ParsedJoints
	for _, skin := range gltfDocument.Skins {
		parsedJoints, err = parseJoints(gltfDocument, skin)
		if err != nil {
			return nil, err
		}
	}

	rootParentTransform := mgl32.Ident4()
	if parsedJoints != nil {
		rootParentTransform = rootParentTransforms(gltfDocument, parsedJoints)
		parsedJoints.RootJoint.LocalBindTransform = rootParentTransform.Mul4(parsedJoints.RootJoint.LocalBindTransform)
		document.JointMap = parsedJoints.JointMap

		parsedAnimations := map[string]*modelspec.AnimationSpec{}
		for _, animation := range gltfDocument.Animations {
			parsedAnimation, err := parseAnimation(ctx, gltfDocument, animation, parsedJoints, rootParentTransform)
			if err != nil {
				return nil, err
			}
			parsedAnimations[animation.Name] = parsedAnimation
		}
		document.Animations = parsedAnimations
	}

	for _, texture := range gltfDocument.Textures {
		img := gltfDocument.Images[int(*texture.Source)]

		if img.MimeType == "" {
			extension := path.Ext(img.URI)
			if extension != ".png" {
				panic(fmt.Sprintf("image extension %s is not supported (png, jpeg, jpg)", extension))
			}
		} else if img.MimeType != "image/png" && img.MimeType != "image/jpeg" && img.MimeType != "image/jpg" {
			panic(fmt.Sprintf("image %s has mimetype %s which is not supported for textures (png, jpeg, jpg)", img.Name, img.MimeType))
		}

		name := strings.Split(img.URI, ".")[0]
		document.Textures = append(document.Textures, name)
	}

	materialSpecs, materialIndexMapping, err := parseMaterialSpecs(gltfDocument, document.Textures)
	if err != nil {
		return nil, err
	}
	document.Materials = materialSpecs

	for i, mesh := range gltfDocument.Meshes {
		primitiveSpecs, err := parsePrimitiveSpecs(gltfDocument, mesh, materialIndexMapping, config)
		if err != nil {
			return nil, err
		}

		meshSpec := &modelspec.Mesh{ID: i, Primitives: primitiveSpecs}
		document.Meshes = append(document.Meshes, meshSpec)
	}

	for _, docScene := range gltfDocument.Scenes {
		scene := modelspec.Scene{}
		for _, node := range docScene.Nodes {
			scene.Nodes = append(scene.Nodes, parseNode(gltfDocument, node))
		}
		document.Scenes = append(document.Scenes, &scene)
		// only load one scene for now
		break
	}

	if parsedJoints != nil {
		document.RootJoint = parsedJoints.RootJoint
	}

	return &document, nil
}

func GetPeripheralFiles(documentPath string) ([]string, error) {
	gltfDocument, err := gltf.Open(documentPath)
	if err != nil {
		return nil, err
	}

	return getPeripheralFiles(gltfDocument)
}

func getPeripheralFiles(gltfDocument *gltf.Document) ([]string, error) {
	var peripheralFiles []string

	// set Peripheral files as they're used and copied over for importing
	for _, buffer := range gltfDocument.Buffers {
		if buffer.URI != "" {
			peripheralFiles = append(peripheralFiles, buffer.URI)
		}
	}

	for _, image := range gltfDocument.Images {
		if image.URI != "" {
			peripheralFiles = append(peripheralFiles, image.URI)
		}
	}

	return peripheralFiles, nil
}

func parseNode(document *gltf.Document, root uint32) *modelspec.Node {
	docNode := document.Nodes[int(root)]

	node := &modelspec.Node{
		Name: docNode.Name,
	}
	if docNode.Mesh != nil {
		id := int(*docNode.Mesh)
		node.MeshID = &id
	}

	translation := docNode.Translation
	rotation := docNode.Rotation
	scale := docNode.Scale

	node.Translation = translation
	node.Rotation = mgl32.Quat{V: mgl32.Vec3{rotation[0], rotation[1], rotation[2]}, W: rotation[3]}
	node.Scale = scale
	node.Transform = getNodeTransform(docNode)

	for _, childNodeID := range docNode.Children {
		node.Children = append(node.Children, parseNode(document, childNodeID))
	}

	return node
}

func rootParentTransforms(document *gltf.Document, parsedJoints *ParsedJoints) mgl32.Mat4 {
	parents := map[int]*int{}
	for i, node := range document.Nodes {
		nodeID := i
		for _, c := range uint32SliceToIntSlice(node.Children) {
			// take address of loop index
			parents[c] = &nodeID
		}
	}

	transform := mgl32.Ident4()
	node := parsedJoints.JointIDToNodeID[parsedJoints.RootJoint.ID]
	parent := parents[node]
	for parent != nil {
		parentNode := document.Nodes[*parent]
		nodeTransform := getNodeTransform(parentNode)
		transform = nodeTransform.Mul4(transform)
		parent = parents[*parent]
	}

	return transform
}

func getNodeTransform(node *gltf.Node) mgl32.Mat4 {
	translation := node.Translation
	rotation := node.Rotation
	scale := node.Scale

	translationMatrix := mgl32.Translate3D(translation[0], translation[1], translation[2])
	rotationMatrix := mgl32.Quat{V: mgl32.Vec3{rotation[0], rotation[1], rotation[2]}, W: rotation[3]}.Mat4()
	scaleMatrix := mgl32.Scale3D(scale[0], scale[1], scale[2])

	return translationMatrix.Mul4(rotationMatrix.Mul4(scaleMatrix))
}

// parseMaterialSpecs creates MaterialSpecifications from the gltf materials list
// we also return an id mapping from the gltf id to the internal material id
// (this might be overkill since their ids are probably also zero index and incrementing)
func parseMaterialSpecs(document *gltf.Document, textures []string) ([]modelspec.Material, map[int]string, error) {
	var materials []modelspec.Material
	idMapping := map[int]string{}

	for gltfIdx, gltfMaterial := range document.Materials {
		pbr := *gltfMaterial.PBRMetallicRoughness

		alphaMode := modelspec.AlphaModeOpaque
		switch gltfMaterial.AlphaMode {
		case gltf.AlphaMask:
			alphaMode = modelspec.AlphaModeMask
		case gltf.AlphaBlend:
			alphaMode = modelspec.AlphaModeBlend
		}

		pbrMaterial := modelspec.PBRMaterial{
			PBRMetallicRoughness: modelspec.PBRMetallicRoughness{
				BaseColorFactor: mgl32.Vec4{pbr.BaseColorFactor[0], pbr.BaseColorFactor[1], pbr.BaseColorFactor[2], pbr.BaseColorFactor[3]},
				MetalicFactor:   *pbr.MetallicFactor,
				RoughnessFactor: *pbr.RoughnessFactor,
			},
			AlphaMode: alphaMode,
		}

		if pbr.BaseColorTexture != nil {
			var intIndex int = int(pbr.BaseColorTexture.Index)
			pbrMaterial.PBRMetallicRoughness.BaseColorTextureIndex = intIndex
			pbrMaterial.PBRMetallicRoughness.BaseColorTextureName = textures[intIndex]
			pbrMaterial.PBRMetallicRoughness.BaseColorTextureCoordsIndex = int(pbr.BaseColorTexture.TexCoord)
		}
		material := modelspec.Material{
			ID:          fmt.Sprintf("%d", gltfIdx),
			PBRMaterial: pbrMaterial,
		}

		idMapping[gltfIdx] = fmt.Sprintf("%d", len(materials))
		materials = append(materials, material)
	}
	return materials, idMapping, nil
}

// parsePrimitiveSpecs takes a gltf mesh and creates a primitive spec for each primitive within the mesh
// index - the index of the mesh, since meshes can have multiple primitives, we can have
// mesh model specifications with the same index. this is okay, external applications should
// not reference this and instead use the mesh id
func parsePrimitiveSpecs(document *gltf.Document, mesh *gltf.Mesh, materialIndexMapping map[int]string, config *ParseConfig) ([]*modelspec.Primitive, error) {
	var primitiveSpecs []*modelspec.Primitive

	for _, primitive := range mesh.Primitives {
		if primitive.Mode != gltf.PrimitiveTriangles {
			panic(fmt.Sprintf("received primitive mode %v but can only support triangles", primitive.Mode))
		}
		primitiveSpec := &modelspec.Primitive{}
		acrIndex := *primitive.Indices
		indicesAccessor := document.Accessors[int(acrIndex)]
		meshIndices, err := modeler.ReadIndices(document, indicesAccessor, nil)
		if err != nil {
			return nil, err
		}
		primitiveSpec.VertexIndices = meshIndices

		if primitive.Material != nil {
			gltfMaterialIndex := int(*primitive.Material)
			primitiveSpec.MaterialIndex = materialIndexMapping[gltfMaterialIndex]
		}

		for attribute, index := range primitive.Attributes {
			acr := document.Accessors[int(index)]
			if primitiveSpec.UniqueVertices == nil {
				primitiveSpec.UniqueVertices = make([]modelspec.Vertex, int(acr.Count))
			}

			if attribute == gltf.POSITION {
				positions, err := modeler.ReadPosition(document, acr, nil)
				if err != nil {
					return nil, err
				}

				for i, position := range positions {
					primitiveSpec.UniqueVertices[i].Position = position
				}
			} else if attribute == gltf.NORMAL {
				normals, err := modeler.ReadNormal(document, acr, nil)
				if err != nil {
					return nil, err
				}
				for i, normal := range normals {
					primitiveSpec.UniqueVertices[i].Normal = normal
				}
			} else if attribute == gltf.TEXCOORD_0 {
				textureCoords, err := modeler.ReadTextureCoord(document, acr, nil)
				if err != nil {
					return nil, err
				}
				for i, textureCoord := range textureCoords {
					if config.TextureCoordStyle == TextureCoordStyleOpenGL {
						textureCoord[1] = 1 - textureCoord[1]
					}
					primitiveSpec.UniqueVertices[i].Texture0Coords = textureCoord
				}
			} else if attribute == gltf.TEXCOORD_1 {
				textureCoords, err := modeler.ReadTextureCoord(document, acr, nil)
				if err != nil {
					return nil, err
				}
				for i, textureCoord := range textureCoords {
					if config.TextureCoordStyle == TextureCoordStyleOpenGL {
						textureCoord[1] = 1 - textureCoord[1]
					}
					primitiveSpec.UniqueVertices[i].Texture1Coords = textureCoord
				}
			} else if attribute == gltf.JOINTS_0 {
				jointsSlice, err := modeler.ReadJoints(document, acr, nil)
				if err != nil {
					return nil, err
				}
				readJointIDs := loosenUint16Array(jointsSlice)
				for i, jointIDs := range readJointIDs {
					primitiveSpec.UniqueVertices[i].JointIDs = jointIDs
				}
			} else if attribute == gltf.WEIGHTS_0 {
				weights, err := modeler.ReadWeights(document, acr, nil)
				if err != nil {
					return nil, err
				}
				readJointWeights := loosenFloat32Array4(weights)
				for i, jointWeights := range readJointWeights {
					primitiveSpec.UniqueVertices[i].JointWeights = jointWeights
				}
			}
		}

		for _, index := range primitiveSpec.VertexIndices {
			primitiveSpec.Vertices = append(primitiveSpec.Vertices, primitiveSpec.UniqueVertices[index])
		}
		primitiveSpecs = append(primitiveSpecs, primitiveSpec)
	}

	return primitiveSpecs, nil
}

func loosenFloat32Array4(floats [][4]float32) [][]float32 {
	result := make([][]float32, len(floats))
	for i, children := range floats {
		result[i] = make([]float32, len(children))
		for j, float := range children {
			result[i][j] = float
		}
	}
	return result
}

func loosenUint16Array(uints [][4]uint16) [][]int {
	result := make([][]int, len(uints))
	for i, children := range uints {
		result[i] = make([]int, len(children))
		for j, uint := range children {
			result[i][j] = int(uint)
		}
	}
	return result
}

func uint32SliceToIntSlice(slice []uint32) []int {
	var result []int
	for _, value := range slice {
		result = append(result, int(value))
	}
	return result
}
