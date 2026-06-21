package gltf

import (
	"fmt"
	"sort"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/kkevinchou/izzet/internal/modelspec"
	"github.com/kkevinchou/izzet/internal/utils"
	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"
)

type jointMeta struct {
	inverseBindMatrix mgl32.Mat4
}

type ParsedJoints struct {
	RootJoint       *modelspec.Joint
	NodeIDToJointID map[int]int
	JointIDToNodeID map[int]int
	JointMap        map[int]*modelspec.Joint
}

type preparedAnimationChannel struct {
	jointID             int
	path                gltf.TRSProperty
	timestamps          []float32
	outputAccessorIndex int
}

func parseAnimation(ctx *parseContext, document *gltf.Document, animation *gltf.Animation, parsedJoints *ParsedJoints, rootParentTransform mgl32.Mat4) (*modelspec.AnimationSpec, error) {
	pTranslation, pRotation, pScale := utils.Decompose(rootParentTransform)

	channels, allTimestamps, err := prepareAnimationChannels(ctx, document, animation, parsedJoints)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare animation channels: [%w]", err)
	}

	jointCount := len(parsedJoints.JointMap)
	keyFrames := make([]*modelspec.KeyFrame, len(allTimestamps))
	for i, timestamp := range allTimestamps {
		pose := make([]modelspec.JointTransform, jointCount)
		defaultTransform := modelspec.NewDefaultJointTransform()
		for jointID := range pose {
			pose[jointID] = defaultTransform
		}
		if parsedJoints.RootJoint.ID >= 0 && parsedJoints.RootJoint.ID < len(pose) {
			// in gltf files, the root joint can be affected by parent nodes that aren't part of the skin.
			// their transformations still apply, so we bake it into the root joint.
			pose[parsedJoints.RootJoint.ID] = modelspec.NewJointTransform(pTranslation, pRotation, pScale)
		}

		// hacky way to keep precision on tiny fractional seconds
		keyFrames[i] = &modelspec.KeyFrame{
			Start: time.Duration(timestamp*1000) * time.Millisecond,
			Pose:  pose,
		}
	}

	for _, channel := range channels {
		timestamps := channel.timestamps
		if timestamps[0] != allTimestamps[0] {
			return nil, fmt.Errorf("first timestamp for channel doesn't match first of all collected timestamps")
		}

		// usage of cursors in the following code is meant to support "optimized" animations
		// for example, in Blender, if a transform is the same between one frame to the next
		// it is simply not set. this can be detected when the output count does not match
		// our collected timestamp count. to handle optimized animations, we simply advance
		// a cursor that points to the current channel's outputs. if we see a timestamp
		// that the output cursor is not aware of (since it was optimized away) we copy the
		// last keyframe's transform
		outputAccessor := document.Accessors[channel.outputAccessorIndex]
		if channel.path == gltf.TRSTranslation {
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
				if localCursor >= len(timestamps) || timestamps[localCursor] != ts {
					pose := keyFrames[i].Pose[channel.jointID]
					pose.Translation = keyFrames[i-1].Pose[channel.jointID].Translation
					keyFrames[i].Pose[channel.jointID] = pose
				} else {
					f32Output := f32OutputValues[localCursor]
					pose := keyFrames[i].Pose[channel.jointID]
					pose.Translation = mgl32.Vec3{f32Output[0], f32Output[1], f32Output[2]}
					keyFrames[i].Pose[channel.jointID] = pose
					localCursor++
				}
			}
		} else if channel.path == gltf.TRSRotation {
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
				if localCursor >= len(timestamps) || timestamps[localCursor] != ts {
					pose := keyFrames[i].Pose[channel.jointID]
					pose.Rotation = keyFrames[i-1].Pose[channel.jointID].Rotation
					keyFrames[i].Pose[channel.jointID] = pose
				} else {
					f32Output := f32OutputValues[localCursor]
					pose := keyFrames[i].Pose[channel.jointID]
					pose.Rotation = mgl32.Quat{V: mgl32.Vec3{f32Output[0], f32Output[1], f32Output[2]}, W: f32Output[3]}
					keyFrames[i].Pose[channel.jointID] = pose
					localCursor++
				}
			}
		} else if channel.path == gltf.TRSScale {
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
				if localCursor >= len(timestamps) || timestamps[localCursor] != ts {
					pose := keyFrames[i].Pose[channel.jointID]
					pose.Scale = keyFrames[i-1].Pose[channel.jointID].Scale
					keyFrames[i].Pose[channel.jointID] = pose
				} else {
					f32Output := f32OutputValues[localCursor]
					pose := keyFrames[i].Pose[channel.jointID]
					pose.Scale = mgl32.Vec3{f32Output[0], f32Output[1], f32Output[2]}
					keyFrames[i].Pose[channel.jointID] = pose
					localCursor++
				}
			}
		}
	}

	return &modelspec.AnimationSpec{
		Name:      animation.Name,
		KeyFrames: keyFrames,
		Length:    keyFrames[len(keyFrames)-1].Start,
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

	joints := map[int]*modelspec.Joint{}
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

		joints[jointID] = &modelspec.Joint{
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
		}
	}

	root := selectRootJoint(joints, jointNodeIDs, nodeIDToJointID, childIDSet)

	parsedJoints := &ParsedJoints{
		RootJoint:       root,
		NodeIDToJointID: nodeIDToJointID,
		JointIDToNodeID: jointIDToNodeID,
		JointMap:        joints,
	}
	return parsedJoints, nil
}

func prepareAnimationChannels(ctx *parseContext, document *gltf.Document, animation *gltf.Animation, parsedJoints *ParsedJoints) ([]preparedAnimationChannel, []float32, error) {
	var preparedChannels []preparedAnimationChannel
	allTimestamps := map[float32]bool{}

	for _, channel := range animation.Channels {
		if channel.Target.Node == nil {
			continue
		}
		nodeID := int(*channel.Target.Node)
		if _, ok := parsedJoints.NodeIDToJointID[nodeID]; !ok {
			continue
		}

		jointID := parsedJoints.NodeIDToJointID[nodeID]
		if channel.Sampler == nil {
			return nil, nil, fmt.Errorf("animation channel for joint %d has no sampler", jointID)
		}

		sampler := animation.Samplers[(*channel.Sampler)]
		if sampler.Interpolation != gltf.InterpolationStep && sampler.Interpolation != gltf.InterpolationLinear {
			return nil, nil, fmt.Errorf("unsupported interpolation \"%s\" found. only \"%s\" and \"%s\" interpolations are supported",
				sampler.Interpolation.String(),
				gltf.InterpolationStep.String(),
				gltf.InterpolationLinear.String(),
			)
		}

		timestamps, err := ctx.readFloatScalarAccessor(document, int(sampler.Input))
		if err != nil {
			return nil, nil, err
		}
		if len(timestamps) == 0 {
			return nil, nil, fmt.Errorf("animation channel for joint %d has no timestamps", jointID)
		}

		for _, timestamp := range timestamps {
			allTimestamps[timestamp] = true
		}

		preparedChannels = append(preparedChannels, preparedAnimationChannel{
			jointID:             jointID,
			path:                channel.Target.Path,
			timestamps:          timestamps,
			outputAccessorIndex: int(sampler.Output),
		})
	}

	var timestamps []float32
	for ts := range allTimestamps {
		timestamps = append(timestamps, ts)
	}

	sort.Slice(timestamps, func(i, j int) bool { return timestamps[i] < timestamps[j] })
	return preparedChannels, timestamps, nil
}

// a skin requires its joints to share a common parent node known as the root.
// that root may or may not be a joint (i.e. listed under skin.joints)
// if skin.skeleton is present, it points to the root node
//
// TODO - support multiple skins
func selectRootJoint(joints map[int]*modelspec.Joint, jointNodeIDs []int, nodeIDToJointID map[int]int, childIDSet map[int]bool) *modelspec.Joint {
	for _, nodeID := range jointNodeIDs {
		jointID := nodeIDToJointID[nodeID]
		if childIDSet[jointID] {
			continue
		}

		joint := joints[jointID]
		if len(joint.Children) > 0 {
			return joint
		}
	}

	for _, nodeID := range jointNodeIDs {
		jointID := nodeIDToJointID[nodeID]
		if !childIDSet[jointID] {
			return joints[jointID]
		}
	}

	return nil
}
