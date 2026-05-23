package animationv2

import (
	"fmt"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/kkevinchou/izzet/internal/modelspec"
	"github.com/kkevinchou/izzet/internal/utils"
)

type AnimationPlayer struct {
	elapsedTime         time.Duration
	animationTransforms map[int]mgl32.Mat4
	currentAnimation    *modelspec.AnimationSpec

	// these fields are from the loaded animation and should not be modified
	animations map[string]*modelspec.AnimationSpec
	rootJoint  *modelspec.JointSpec
}

// support blend tree operations
func NewAnimationPlayer() *AnimationPlayer {
	return &AnimationPlayer{}
}

func (player *AnimationPlayer) Initialize(animations map[string]*modelspec.AnimationSpec, rootJoint *modelspec.JointSpec) {
	player.animations = animations
	player.rootJoint = rootJoint
}

func (player *AnimationPlayer) NormalizedClipProgress() float64 {
	return float64(player.elapsedTime.Milliseconds() / player.currentAnimation.Length.Milliseconds())
}

func (player *AnimationPlayer) CurrentAnimation() string {
	if player.currentAnimation == nil {
		return ""
	}
	return player.currentAnimation.Name
}

func (player *AnimationPlayer) AnimationTransforms() map[int]mgl32.Mat4 {
	return player.animationTransforms
}

func (player *AnimationPlayer) BindPoseTransforms() map[int]mgl32.Mat4 {
	animationTransforms := map[int]mgl32.Mat4{}
	bindPoseHelper(player.rootJoint, animationTransforms)
	return animationTransforms
}

func (player *AnimationPlayer) PlayClip(animationName string) {
	if currentAnimation, ok := player.animations[animationName]; ok {
		player.currentAnimation = currentAnimation
		player.elapsedTime = 0
	} else {
		panic(fmt.Sprintf("failed to find animation %s", animationName))
	}
}

func (player *AnimationPlayer) Length() time.Duration {
	if player.currentAnimation == nil {
		return 0
	}
	return player.currentAnimation.Length
}

func (player *AnimationPlayer) Update(delta time.Duration) {
	if player.currentAnimation == nil {
		return
	}

	player.elapsedTime += delta
	for player.elapsedTime > player.currentAnimation.Length {
		player.elapsedTime = player.currentAnimation.Length
	}

	pose := calculateCurrentAnimationPose(player.elapsedTime, player.currentAnimation.KeyFrames)
	poseTransforms := convertPoseToTransformMatrix(pose)
	animationTransforms := map[int]mgl32.Mat4{}
	computeJointTransformsHelper(player.rootJoint, mgl32.Ident4(), poseTransforms, animationTransforms)
	player.animationTransforms = animationTransforms
}

func computeJointTransformsHelper(joint *modelspec.JointSpec, parentTransform mgl32.Mat4, poseTransforms map[int]mgl32.Mat4, transforms map[int]mgl32.Mat4) {
	localTransform := poseTransforms[joint.ID]

	if _, ok := poseTransforms[joint.ID]; !ok {
		// if there is no pose in the animation, just use the local bind transform.
		// when a pose is present, the keyframe data already includes the local bind transform

		// using the identity matrix is not correct, we still need the bind transform.
		// localTransform = mgl32.Ident4()

		// further down, we multiply by the inverse bind transform which is what allows us
		// to cancel out the local bind transform and produce the actual output identity matrix

		localTransform = joint.LocalBindTransform
	}

	// model-space transform that includes all the parental transforms
	// and the local transform and is not meant to be used to transform any vertices
	// until we multiply it by the inverse bind transform
	poseTransform := parentTransform.Mul4(localTransform)

	for _, child := range joint.Children {
		computeJointTransformsHelper(child, poseTransform, poseTransforms, transforms)
	}

	// this is the model-space transform that can finally be used to transform
	// any vertices it influences
	transforms[joint.ID] = poseTransform.Mul4(joint.InverseBindTransform)
}

func calculateCurrentAnimationPose(elapsedTime time.Duration, keyFrames []*modelspec.KeyFrame) map[int]modelspec.JointTransform {
	var startKeyFrame *modelspec.KeyFrame
	var endKeyFrame *modelspec.KeyFrame
	var progression float32

	// iterate backwards looking for the starting keyframe
	for i := len(keyFrames) - 1; i >= 0; i-- {
		var startKeyFrameIndex int
		if elapsedTime >= keyFrames[i].Start {
			startKeyFrameIndex = i
		} else if i == 0 {
			startKeyFrameIndex = len(keyFrames) - 1
		} else {
			continue
		}
		endKeyFrameIndex := (startKeyFrameIndex + 1) % len(keyFrames)

		startKeyFrame = keyFrames[startKeyFrameIndex]
		endKeyFrame = keyFrames[endKeyFrameIndex]
		startKeyFrameTimestamp := startKeyFrame.Start
		if startKeyFrameIndex > endKeyFrameIndex {
			// handle case where we're looping from the last key frame
			startKeyFrameTimestamp = 0
		}
		progression = float32(elapsedTime-startKeyFrameTimestamp) / float32((endKeyFrame.Start - startKeyFrameTimestamp))
		break
	}

	// progression = 0
	// startKeyFrame = keyFrames[0]
	return interpolatePoses(startKeyFrame.Pose, endKeyFrame.Pose, progression)
}

func convertPoseToTransformMatrix(pose map[int]modelspec.JointTransform) map[int]mgl32.Mat4 {
	transformMatrices := map[int]mgl32.Mat4{}
	for jointID, transform := range pose {
		translation := transform.Translation
		rotation := transform.Rotation.Mat4()
		scale := transform.Scale
		transformMatrices[jointID] = mgl32.Translate3D(translation.X(), translation.Y(), translation.Z()).Mul4(rotation).Mul4(mgl32.Scale3D(scale.X(), scale.Y(), scale.Z()))
	}
	return transformMatrices
}

func interpolatePoses(j1, j2 map[int]modelspec.JointTransform, progression float32) map[int]modelspec.JointTransform {
	if progression > 1 {
		progression = 1
	}

	if progression < 0 {
		progression = 0
	}

	interpolatedPose := map[int]modelspec.JointTransform{}
	for jointID := range j1 {
		k1JointTransform := j1[jointID]
		k2JointTransform := j2[jointID]

		// WTF - this lerp doesn't look right when interpolating keyframes???
		// rotationQuat := mgl32.QuatLerp(k1JointTransform.Rotation, k2JointTransform.Rotation, progression)
		rotation := utils.QInterpolate(k1JointTransform.Rotation, k2JointTransform.Rotation, progression)

		translation := k1JointTransform.Translation.Add(k2JointTransform.Translation.Sub(k1JointTransform.Translation).Mul(progression))
		scale := k1JointTransform.Scale.Add(k2JointTransform.Scale.Sub(k1JointTransform.Scale).Mul(progression))

		// interpolatedPose[jointID] = mgl32.Translate3D(translation.X(), translation.Y(), translation.Z()).Mul4(rotation).Mul4(mgl32.Scale3D(scale.X(), scale.Y(), scale.Z()))
		interpolatedPose[jointID] = modelspec.JointTransform{
			Translation: translation,
			Rotation:    rotation,
			Scale:       scale,
		}
	}
	return interpolatedPose
}

func bindPoseHelper(joint *modelspec.JointSpec, transforms map[int]mgl32.Mat4) {
	transforms[joint.ID] = joint.FullBindTransform
	for _, child := range joint.Children {
		bindPoseHelper(child, transforms)
	}
}
