package animation

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

	currentPose     []modelspec.JointTransform
	blendSourcePose []modelspec.JointTransform
	blendDuration   time.Duration

	// these fields are from the loaded animation and should not be modified
	animations map[string]*modelspec.AnimationSpec
	rootJoint  *modelspec.Joint

	leftover time.Duration
	playRate float64
}

// support blend tree operations
func NewAnimationPlayer() *AnimationPlayer {
	return &AnimationPlayer{playRate: 1}
}

func (player *AnimationPlayer) Initialize(animations map[string]*modelspec.AnimationSpec, rootJoint *modelspec.Joint) {
	player.animations = animations
	player.rootJoint = rootJoint
}

func (player *AnimationPlayer) SetPlayRate(rate float64) {
	player.playRate = rate
}

func (player *AnimationPlayer) PlayRate() float64 {
	return player.playRate
}

func (player *AnimationPlayer) NormalizedClipProgress() float64 {
	return float64(player.elapsedTime.Milliseconds()) / float64(player.currentAnimation.Length.Milliseconds())
}

func (player *AnimationPlayer) ElapsedTime() time.Duration {
	return player.elapsedTime
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

var clip string

func (player *AnimationPlayer) PlayClip(clipName string) {
	if currentAnimation, ok := player.animations[clipName]; ok {
		player.currentAnimation = currentAnimation
		player.elapsedTime = 0
		clip = clipName
	} else {
		panic(fmt.Sprintf("failed to find animation %s", clipName))
	}
}

func (player *AnimationPlayer) BlendClip(clipName string, blendDuration time.Duration) {
	player.PlayClip(clipName)
	player.blendSourcePose = player.currentPose
	player.blendDuration = blendDuration
}

func (player *AnimationPlayer) Update(delta time.Duration) {
	if player.currentAnimation == nil {
		return
	}

	if player.elapsedTime == player.currentAnimation.Length {
		// Update() has been called on the animation player while the current clip has finished playing.
		// skip incrementing elapsedTime and leftover to avoid runaway time values.
		// the animation transforms are for the final frame of the current clip
		player.leftover = 0
		return
	}

	player.elapsedTime += time.Duration(float64(delta+player.leftover) * player.playRate)
	player.leftover = 0

	if player.elapsedTime > player.currentAnimation.Length {
		overrunClipTime := player.elapsedTime - player.currentAnimation.Length
		player.leftover = time.Duration(float64(overrunClipTime) / player.playRate)
		player.elapsedTime = player.currentAnimation.Length
	}

	pose := calculateCurrentAnimationPose(player.elapsedTime, player.currentAnimation.KeyFrames)

	if player.blendSourcePose != nil {
		t := float64(player.elapsedTime) / float64(player.blendDuration)
		t = min(1, t)
		pose = interpolatePoses(player.blendSourcePose, pose, float32(t))

		if t == 1 {
			player.blendSourcePose = nil
		}
	}

	player.currentPose = pose
	poseTransforms := convertPoseToTransformMatrix(pose)
	animationTransforms := map[int]mgl32.Mat4{}
	computeJointTransformsHelper(player.rootJoint, mgl32.Ident4(), poseTransforms, animationTransforms)

	player.animationTransforms = animationTransforms
}

func (player *AnimationPlayer) SetCurrentAnimationFrame(animation string, keyframe int) {
	if _, ok := player.animations[animation]; !ok {
		return
	}
	player.currentAnimation = player.animations[animation]
	keyFrames := player.currentAnimation.KeyFrames

	if keyframe >= len(keyFrames) {
		return
	}

	startKeyFrame := keyFrames[keyframe]
	endKeyFrame := keyFrames[(keyframe+1)%len(keyFrames)]

	pose := interpolatePoses(startKeyFrame.Pose, endKeyFrame.Pose, 0)
	poseTransforms := convertPoseToTransformMatrix(pose)
	animationTransforms := map[int]mgl32.Mat4{}
	computeJointTransformsHelper(player.rootJoint, mgl32.Ident4(), poseTransforms, animationTransforms)
	player.animationTransforms = animationTransforms
}

func (player *AnimationPlayer) Length() time.Duration {
	if player.currentAnimation == nil {
		return 0
	}
	return player.currentAnimation.Length
}

func computeJointTransformsHelper(joint *modelspec.Joint, parentTransform mgl32.Mat4, poseTransforms []mgl32.Mat4, transforms map[int]mgl32.Mat4) {
	// if there is no pose in the animation, just use the local bind transform.
	// when a pose is present, the keyframe data already includes the local bind transform
	// using the identity matrix is not correct, we still need the bind transform.

	// further down, we multiply by the inverse bind transform which is what allows us
	// to cancel out the local bind transform and produce the actual output identity matrix
	localTransform := joint.LocalBindTransform
	if joint.ID >= 0 && joint.ID < len(poseTransforms) {
		localTransform = poseTransforms[joint.ID]
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

func calculateCurrentAnimationPose(elapsedTime time.Duration, keyFrames []*modelspec.KeyFrame) []modelspec.JointTransform {
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

		progression = float32(elapsedTime-startKeyFrameTimestamp) / float32((endKeyFrame.Start - startKeyFrameTimestamp))
		break
	}

	return interpolatePoses(startKeyFrame.Pose, endKeyFrame.Pose, progression)
}

func convertPoseToTransformMatrix(pose []modelspec.JointTransform) []mgl32.Mat4 {
	transformMatrices := make([]mgl32.Mat4, len(pose))
	for jointID, transform := range pose {
		translation := transform.Translation
		rotation := transform.Rotation.Mat4()
		scale := transform.Scale
		transformMatrices[jointID] = mgl32.Translate3D(translation.X(), translation.Y(), translation.Z()).Mul4(rotation).Mul4(mgl32.Scale3D(scale.X(), scale.Y(), scale.Z()))
	}
	return transformMatrices
}

func interpolatePoses(j1, j2 []modelspec.JointTransform, progression float32) []modelspec.JointTransform {
	if progression > 1 {
		progression = 1
	}

	if progression < 0 {
		progression = 0
	}

	if len(j1) == 0 {
		return clonePose(j2)
	}
	if len(j2) == 0 {
		return clonePose(j1)
	}

	if len(j1) != len(j2) {
		panic(fmt.Errorf("joint count from two interpolated poses differ (%d and %d)", len(j1), len(j2)))
	}
	jointCount := len(j1)

	interpolatedPose := make([]modelspec.JointTransform, jointCount)
	for jointID := 0; jointID < jointCount; jointID++ {
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

func clonePose(pose []modelspec.JointTransform) []modelspec.JointTransform {
	cloned := make([]modelspec.JointTransform, len(pose))
	copy(cloned, pose)
	return cloned
}

func bindPoseHelper(joint *modelspec.Joint, transforms map[int]mgl32.Mat4) {
	transforms[joint.ID] = joint.FullBindTransform
	for _, child := range joint.Children {
		bindPoseHelper(child, transforms)
	}
}
