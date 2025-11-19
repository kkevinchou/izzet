package gltf

import (
	"fmt"
	"log/slog"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/kkevinchou/izzet/internal/iztlog"
	"github.com/kkevinchou/izzet/internal/modelspec"
	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"
)

type jointMeta struct {
	inverseBindMatrix mgl32.Mat4
}

type ParsedJoints struct {
	RootJoint       *modelspec.JointSpec
	NodeIDToJointID map[int]int
	JointIDToNodeID map[int]int
	JointMap        map[int]*modelspec.JointSpec
}

type TextureCoordStyle int

const (
	TextureCoordStyleOpenGL = 1
)

type ParseConfig struct {
	TextureCoordStyle TextureCoordStyle
}

func ParseGLTF(name string, documentPath string, config *ParseConfig) (*modelspec.Document, error) {
	logger := iztlog.Logger.With("document path", documentPath)
	var document modelspec.Document

	document.Name = name

	// if strings.Contains(name, "dude") {
	// 	documentPath = "C:\\Users\\Kevin\\goprojects\\izzet\\.project\\scene\\content\\dude.gltf"
	// }
	gltfDocument, err := gltf.Open(documentPath)
	if err != nil {
		return nil, err
	}

	// set Peripheral files as they're used and copied over for importing
	for _, buffer := range gltfDocument.Buffers {
		if buffer.URI != "" {
			document.PeripheralFiles = append(document.PeripheralFiles, buffer.URI)
		}
	}

	for _, image := range gltfDocument.Images {
		if image.URI != "" {
			document.PeripheralFiles = append(document.PeripheralFiles, image.URI)
		}
	}

	var parsedJoints *ParsedJoints
	for _, skin := range gltfDocument.Skins {
		parsedJoints, err = parseJoints(gltfDocument, skin)
		if err != nil {
			return nil, err
		}
	}
	if parsedJoints != nil {
		document.JointMap = parsedJoints.JointMap
	}

	parsedAnimations := map[string]*modelspec.AnimationSpec{}
	for _, animation := range gltfDocument.Animations {
		parsedAnimation, err := parseAnimation(gltfDocument, animation, parsedJoints)
		parsedAnimations[animation.Name] = parsedAnimation
		if err != nil {
			return nil, err
		}
	}
	if len(parsedAnimations) > 0 {
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

	materialSpecs, materialIndexMapping, err := parseMaterialSpecs(gltfDocument, document.Textures, logger)
	if err != nil {
		iztlog.Logger.Error(err.Error())
		return nil, err
	}
	document.Materials = materialSpecs

	for i, mesh := range gltfDocument.Meshes {
		primitiveSpecs, err := parsePrimitiveSpecs(gltfDocument, mesh, materialIndexMapping, config)
		if err != nil {
			iztlog.Logger.Error(err.Error())
			return nil, err
		}

		meshSpec := &modelspec.MeshSpecification{ID: i, Primitives: primitiveSpecs}
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

	rootTransforms := mgl32.Ident4()
	if parsedJoints != nil {
		document.RootJoint = parsedJoints.RootJoint
		rootTransforms = rootParentTransforms(gltfDocument, parsedJoints)
		_ = rootTransforms
	}

	return &document, nil
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

	translationMatrix := mgl32.Translate3D(translation[0], translation[1], translation[2])
	rotationMatrix := node.Rotation.Mat4()
	scaleMatrix := mgl32.Scale3D(scale[0], scale[1], scale[2])
	node.Transform = translationMatrix.Mul4(rotationMatrix).Mul4(scaleMatrix)

	for _, childNodeID := range docNode.Children {
		node.Children = append(node.Children, parseNode(document, childNodeID))
	}

	return node
}

func rootParentTransforms(document *gltf.Document, parsedJoints *ParsedJoints) mgl32.Mat4 {
	children := map[int][]int{}
	parents := map[int]*int{}
	for i, node := range document.Nodes {
		nodeID := i
		cs := uint32SliceToIntSlice(node.Children)
		children[i] = cs
		for _, c := range cs {
			// take address of loop index
			parents[c] = &nodeID
		}
	}

	transform := mgl32.Ident4()
	node := parsedJoints.JointIDToNodeID[parsedJoints.RootJoint.ID]
	parent := parents[node]
	for parent != nil {
		parentNode := document.Nodes[*parent]
		translation := parentNode.Translation
		rotation := parentNode.Rotation
		scale := parentNode.Scale

		translationMatrix := mgl32.Translate3D(translation[0], translation[1], translation[2])
		rotationMatrix := mgl32.Quat{V: mgl32.Vec3{rotation[0], rotation[1], rotation[2]}, W: rotation[3]}.Mat4()
		scaleMatrix := mgl32.Scale3D(scale[0], scale[1], scale[2])

		nodeTransform := translationMatrix.Mul4(rotationMatrix.Mul4(scaleMatrix))
		transform = nodeTransform.Mul4(transform)
		parent = parents[*parent]
	}

	// transform = mgl32.Ident4()
	return transform
}

// preprocessAnimations is a preprocessing method that loops through all animation channels
// and collects the timestamps associated with them as well as joints that are affected
// by animations
func preprocessAnimations(document *gltf.Document, animation *gltf.Animation, parsedJoints *ParsedJoints) ([]float32, []int, error) {
	allTimestamps := map[float32]bool{}
	jointIDs := map[int]bool{}

	for _, channel := range animation.Channels {
		nodeID := int(*channel.Target.Node)
		if _, ok := parsedJoints.NodeIDToJointID[nodeID]; !ok {
			continue
		}

		jointID := parsedJoints.NodeIDToJointID[nodeID]
		jointIDs[jointID] = true

		sampler := animation.Samplers[(*channel.Sampler)]
		inputAccessorIndex := int(sampler.Input)

		inputAccessor := document.Accessors[inputAccessorIndex]
		if inputAccessor.ComponentType != gltf.ComponentFloat {
			return nil, nil, fmt.Errorf("unexpected component type %v", inputAccessor.ComponentType)
		}
		if inputAccessor.Type != gltf.AccessorScalar {
			return nil, nil, fmt.Errorf("unexpected accessor type %v", inputAccessor.Type)
		}

		input, err := modeler.ReadAccessor(document, inputAccessor, nil)
		if err != nil {
			panic("WHA")
		}

		timestamps := input.([]float32)
		for _, timestamp := range timestamps {
			allTimestamps[timestamp] = true
		}
	}

	var timestamps []float32
	for ts, _ := range allTimestamps {
		timestamps = append(timestamps, ts)
	}

	var sliceJointIDs []int
	for jointID := range jointIDs {
		sliceJointIDs = append(sliceJointIDs, jointID)
	}

	sort.Slice(timestamps, func(i, j int) bool { return timestamps[i] < timestamps[j] })
	sort.Ints(sliceJointIDs)
	return timestamps, sliceJointIDs, nil
}

func parseAnimation(document *gltf.Document, animation *gltf.Animation, parsedJoints *ParsedJoints) (*modelspec.AnimationSpec, error) {
	keyFrames := map[float32]*modelspec.KeyFrame{}

	allTimestamps, allJointIDs, err := preprocessAnimations(document, animation, parsedJoints)
	if err != nil {
		return nil, fmt.Errorf("failed to collect timestamps: [%w]", err)
	}

	// initialize our initial keyframe datastructure with empty values
	for _, timestamp := range allTimestamps {
		// hacky way to keep precision on tiny fractional seconds
		keyFrames[timestamp] = &modelspec.KeyFrame{
			Start: time.Duration(timestamp*1000) * time.Millisecond,
			Pose:  map[int]modelspec.JointTransform{},
		}
		for _, jointID := range allJointIDs {
			keyFrames[timestamp].Pose[jointID] = modelspec.NewDefaultJointTransform()
		}
	}

	for _, channel := range animation.Channels {
		nodeID := int(*channel.Target.Node)
		if _, ok := parsedJoints.NodeIDToJointID[nodeID]; !ok {
			continue
		}

		jointID := parsedJoints.NodeIDToJointID[nodeID]
		sampler := animation.Samplers[(*channel.Sampler)]
		inputAccessorIndex := int(sampler.Input)
		outputAccessorIndex := int(sampler.Output)
		if sampler.Interpolation != gltf.InterpolationStep && sampler.Interpolation != gltf.InterpolationLinear {
			return nil, fmt.Errorf("unsupported interpolation \"%s\" found. only \"%s\" and \"%s\" interpolations are supported",
				sampler.Interpolation.String(),
				gltf.InterpolationStep.String(),
				gltf.InterpolationLinear.String(),
			)
		}

		inputAccessor := document.Accessors[inputAccessorIndex]
		if inputAccessor.ComponentType != gltf.ComponentFloat {
			return nil, fmt.Errorf("unexpected component type %v", inputAccessor.ComponentType)
		}
		if inputAccessor.Type != gltf.AccessorScalar {
			return nil, fmt.Errorf("unexpected accessor type %v", inputAccessor.Type)
		}

		input, err := modeler.ReadAccessor(document, inputAccessor, nil)
		if err != nil {
			panic("WHA")
		}

		timestamps := input.([]float32)
		if timestamps[0] != allTimestamps[0] {
			panic("first timestamp for channel doesn't match first of all collected timestamps")
		}

		// usage of cursors in the following code is meant to support "optimized" animations
		// for example, in Blender, if a transform is the same between one frame to the next
		// it is simply not set. this can be detected when the output count does not match
		// our collected timestamp count. to handle optimized animations, we simply advance
		// a cursor that points to the current channel's outputs. if we see a timestamp
		// that the output cursor is not aware of (since it was optimized away) we copy the
		// last keyframe's transform
		outputAccessor := document.Accessors[outputAccessorIndex]
		if channel.Target.Path == gltf.TRSTranslation {
			if outputAccessor.ComponentType != gltf.ComponentFloat {
				return nil, fmt.Errorf("unexpected component type %v", outputAccessor.ComponentType)
			}
			if outputAccessor.Type != gltf.AccessorVec3 {
				return nil, fmt.Errorf("unexpected accessor type %v", outputAccessor.Type)
			}
			output, err := modeler.ReadAccessor(document, outputAccessor, nil)
			if err != nil {
				panic("WHA")
			}
			if outputAccessor.Count <= 0 {
				panic("accessor with 0 count")
			}

			f32OutputValues := output.([][3]float32)

			localCursor := 0
			for i, ts := range allTimestamps {
				if timestamps[localCursor] != ts {
					lastTS := allTimestamps[i-1]
					pose := keyFrames[ts].Pose[jointID]
					pose.Translation = keyFrames[lastTS].Pose[jointID].Translation
					keyFrames[ts].Pose[jointID] = pose
				} else {
					f32Output := f32OutputValues[localCursor]
					pose := keyFrames[ts].Pose[jointID]
					pose.Translation = mgl32.Vec3{f32Output[0], f32Output[1], f32Output[2]}
					keyFrames[ts].Pose[jointID] = pose
					localCursor++
				}
			}
		} else if channel.Target.Path == gltf.TRSRotation {
			if outputAccessor.ComponentType != gltf.ComponentFloat {
				return nil, fmt.Errorf("unexpected component type %v", outputAccessor.ComponentType)
			}
			if outputAccessor.Type != gltf.AccessorVec4 {
				return nil, fmt.Errorf("unexpected accessor type %v", outputAccessor.Type)
			}
			output, err := modeler.ReadAccessor(document, outputAccessor, nil)
			if err != nil {
				panic("WHA")
			}
			if outputAccessor.Count <= 0 {
				panic("accessor with 0 count")
			}

			f32OutputValues := output.([][4]float32)

			localCursor := 0
			for i, ts := range allTimestamps {
				if timestamps[localCursor] != ts {
					lastTS := allTimestamps[i-1]
					pose := keyFrames[ts].Pose[jointID]
					pose.Rotation = keyFrames[lastTS].Pose[jointID].Rotation
					keyFrames[ts].Pose[jointID] = pose
				} else {
					f32Output := f32OutputValues[localCursor]
					pose := keyFrames[ts].Pose[jointID]
					pose.Rotation = mgl32.Quat{V: mgl32.Vec3{f32Output[0], f32Output[1], f32Output[2]}, W: f32Output[3]}
					keyFrames[ts].Pose[jointID] = pose
					localCursor++
				}
			}
		} else if channel.Target.Path == gltf.TRSScale {
			if outputAccessor.ComponentType != gltf.ComponentFloat {
				return nil, fmt.Errorf("unexpected component type %v", outputAccessor.ComponentType)
			}
			if outputAccessor.Type != gltf.AccessorVec3 {
				return nil, fmt.Errorf("unexpected accessor type %v", outputAccessor.Type)
			}
			output, err := modeler.ReadAccessor(document, outputAccessor, nil)
			if err != nil {
				panic("WHA")
			}
			if outputAccessor.Count <= 0 {
				panic("accessor with 0 count")
			}

			f32OutputValues := output.([][3]float32)

			localCursor := 0
			for i, ts := range allTimestamps {
				// local cursor is ahead, backfill the transform
				if timestamps[localCursor] != ts {
					lastTS := allTimestamps[i-1]
					pose := keyFrames[ts].Pose[jointID]
					pose.Scale = keyFrames[lastTS].Pose[jointID].Scale
					keyFrames[ts].Pose[jointID] = pose
				} else {
					f32Output := f32OutputValues[localCursor]
					pose := keyFrames[ts].Pose[jointID]
					pose.Scale = mgl32.Vec3{f32Output[0], f32Output[1], f32Output[2]}
					keyFrames[ts].Pose[jointID] = pose
					localCursor++
				}
			}
		}
	}

	var keyFrameSlice []*modelspec.KeyFrame
	for _, timestamp := range allTimestamps {
		keyFrameSlice = append(keyFrameSlice, keyFrames[timestamp])
	}

	return &modelspec.AnimationSpec{
		Name:      animation.Name,
		KeyFrames: keyFrameSlice,
		Length:    keyFrameSlice[len(keyFrameSlice)-1].Start,
	}, nil
}

func parseJoints(document *gltf.Document, skin *gltf.Skin) (*ParsedJoints, error) {
	jms := map[int]*jointMeta{}
	jointNodeIDs := uint32SliceToIntSlice(skin.Joints)
	nodeIDToJointID := map[int]int{}
	jointIDToNodeID := map[int]int{}

	for jointID, nodeID := range jointNodeIDs {
		nodeIDToJointID[nodeID] = jointID
		jointIDToNodeID[jointID] = nodeID
	}

	for jointID := 0; jointID < len(jointNodeIDs); jointID++ {
		jms[jointID] = &jointMeta{}
	}

	acr := document.Accessors[int(*skin.InverseBindMatrices)]
	if acr.ComponentType != gltf.ComponentFloat {
		return nil, fmt.Errorf("unexpected component type %v", acr.ComponentType)
	}
	if acr.Type != gltf.AccessorMat4 {
		return nil, fmt.Errorf("unexpected accessor type %v", acr.Type)
	}

	data, err := modeler.ReadAccessor(document, acr, nil)
	if err != nil {
		return nil, err
	}

	inverseBindMatrices := data.([][4][4]float32)
	for jointID, _ := range jms {
		matrix := inverseBindMatrices[jointID]
		inverseBindMatrix := mgl32.Mat4FromRows(
			mgl32.Vec4{matrix[0][0], matrix[0][1], matrix[0][2], matrix[0][3]},
			mgl32.Vec4{matrix[1][0], matrix[1][1], matrix[1][2], matrix[1][3]},
			mgl32.Vec4{matrix[2][0], matrix[2][1], matrix[2][2], matrix[2][3]},
			mgl32.Vec4{matrix[3][0], matrix[3][1], matrix[3][2], matrix[3][3]},
		)
		jms[jointID].inverseBindMatrix = inverseBindMatrix
	}

	joints := map[int]*modelspec.JointSpec{}
	for nodeID, node := range document.Nodes {
		if _, ok := nodeIDToJointID[nodeID]; !ok {
			continue
		}

		jointID := nodeIDToJointID[nodeID]

		translation := node.Translation
		rotation := node.Rotation
		scale := node.Scale
		// from the gltf spec:
		//
		// When a node is targeted for animation (referenced by an animation.channel.target),
		// only TRS properties MAY be present; matrix MUST NOT be present.
		translationMatrix := mgl32.Translate3D(translation[0], translation[1], translation[2])
		rotationMatrix := mgl32.Quat{V: mgl32.Vec3{rotation[0], rotation[1], rotation[2]}, W: rotation[3]}.Mat4()
		scaleMatrix := mgl32.Scale3D(scale[0], scale[1], scale[2])

		joints[jointID] = &modelspec.JointSpec{
			Name:                 fmt.Sprintf("joint_%s_%d", node.Name, jointID),
			ID:                   jointID,
			LocalBindTransform:   translationMatrix.Mul4(rotationMatrix.Mul4(scaleMatrix)),
			InverseBindTransform: jms[jointID].inverseBindMatrix,
			FullBindTransform:    jms[jointID].inverseBindMatrix.Inv(),
		}
	}

	// set up the joint hierarchy
	childIDSet := map[int]bool{}
	for jointID, nodeID := range jointNodeIDs {
		children := uint32SliceToIntSlice(document.Nodes[nodeID].Children)
		// there can be children that aren't joints, need to make an explicit check
		for _, childNodeID := range children {
			if _, ok := nodeIDToJointID[childNodeID]; !ok {
				continue
			}
			childJointID := nodeIDToJointID[childNodeID]
			childIDSet[childJointID] = true
			joints[jointID].Children = append(joints[jointID].Children, joints[childJointID])
			joints[childJointID].Parent = joints[jointID]
		}
	}

	// find the root
	var root *modelspec.JointSpec
	for id, _ := range joints {
		if _, ok := childIDSet[id]; !ok {
			joint := joints[id]
			if len(joint.Children) > 0 {
				// sometimes people put joints as control objects that aren't actual parents
				root = joint
				break
			}
		}
	}

	parsedJoints := &ParsedJoints{
		RootJoint:       root,
		NodeIDToJointID: nodeIDToJointID,
		JointIDToNodeID: jointIDToNodeID,
		JointMap:        joints,
	}
	return parsedJoints, nil
}

// parseMaterialSpecs creates MaterialSpecifications from the gltf materials list
// we also return an id mapping from the gltf id to the internal material id
// (this might be overkill since their ids are probably also zero index and incrementing)
func parseMaterialSpecs(document *gltf.Document, textures []string, logger *slog.Logger) ([]modelspec.MaterialSpecification, map[int]string, error) {
	var materials []modelspec.MaterialSpecification
	idMapping := map[int]string{}

	for gltfIdx, gltfMaterial := range document.Materials {
		pbr := *gltfMaterial.PBRMetallicRoughness

		alphaMode := modelspec.AlphaModeOpaque
		switch gltfMaterial.AlphaMode {
		case gltf.AlphaMask:
			alphaMode = modelspec.AlphaModeMask
			logger.Warn("unsupported alpha mode alpha mask")
		case gltf.AlphaBlend:
			alphaMode = modelspec.AlphaModeBlend
			logger.Warn("unsupported alpha mode alpha blend")
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
		material := modelspec.MaterialSpecification{
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
func parsePrimitiveSpecs(document *gltf.Document, mesh *gltf.Mesh, materialIndexMapping map[int]string, config *ParseConfig) ([]*modelspec.PrimitiveSpecification, error) {
	var primitiveSpecs []*modelspec.PrimitiveSpecification

	if len(mesh.Primitives) > 1 {
		iztlog.Logger.Info(mesh.Name)
	}

	for _, primitive := range mesh.Primitives {
		if primitive.Mode != gltf.PrimitiveTriangles {
			panic(fmt.Sprintf("received primitive mode %v but can only support triangles", primitive.Mode))
		}
		primitiveSpec := &modelspec.PrimitiveSpecification{}
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

				if len(positions) != len(primitiveSpec.UniqueVertices) {
					iztlog.Logger.Info("dafuq")
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
			} else {
				iztlog.Logger.Info("[%s] unhandled attribute %s\n", mesh.Name, attribute)
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

func loosenFloat32Array2ToVec(floats [][2]float32) []mgl32.Vec2 {
	var result []mgl32.Vec2
	for _, props := range floats {
		result = append(result, mgl32.Vec2(props))
	}
	return result
}

func loosenFloat32Array3ToVec(floats [][3]float32) []mgl32.Vec3 {
	var result []mgl32.Vec3
	for _, props := range floats {
		result = append(result, mgl32.Vec3(props))
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
