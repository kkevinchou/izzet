package animation

import (
	"fmt"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/kkevinchou/izzet/lib/libutils"
	"github.com/kkevinchou/izzet/lib/model"
	"github.com/kkevinchou/izzet/lib/modelspec"
)

type AnimationPlayer struct {
	elapsedTime         time.Duration
	animationTransforms map[int]mgl32.Mat4
	currentAnimation    *modelspec.AnimationSpec

	// these fields are from the loaded animation and should not be modified
	animations map[string]*modelspec.AnimationSpec
	rootJoint  *modelspec.JointSpec

	secondaryAnimation *string
	loop               bool

	blendActive               bool
	blendAnimationElapsedTime time.Duration
	blendAnimation            *modelspec.AnimationSpec
	blendDuration             time.Duration
	blendDurationSoFar        time.Duration
}

// func NewAnimationPlayer(animations map[string]*modelspec.AnimationSpec, rootJoint *modelspec.JointSpec) *AnimationPlayer {
func NewAnimationPlayer(m *model.Model) *AnimationPlayer {
	return &AnimationPlayer{
		animations: m.Animations(),
		rootJoint:  m.RootJoint(),
		loop:       true,
	}
}

func (player *AnimationPlayer) CurrentAnimation() string {
	if player.currentAnimation == nil {
		return ""
	}
	if player.blendAnimation != nil {
		return player.blendAnimation.Name
	}
	return player.currentAnimation.Name
}

func (player *AnimationPlayer) AnimationTransforms() map[int]mgl32.Mat4 {
	return player.animationTransforms
}

func (player *AnimationPlayer) PlayAnimation(animationName string) {
	if player.blendAnimation == nil && player.currentAnimation != nil && player.currentAnimation.Name == animationName {
		return
	}
	if player.secondaryAnimation != nil && animationName == *player.secondaryAnimation {
		return
	}

	if currentAnimation, ok := player.animations[animationName]; ok {
		player.currentAnimation = currentAnimation
		player.elapsedTime = 0
		player.blendAnimation = nil
	} else {
		panic(fmt.Sprintf("failed to find animation %s", animationName))
	}
}

func (player *AnimationPlayer) PlayAndBlendAnimation(animationName string, blendDuration time.Duration) {
	if player.currentAnimation == nil {
		fmt.Println("NO ANIMATION TO BLEND FROM")
		return
	}

	if player.blendAnimation == nil && player.currentAnimation != nil && player.currentAnimation.Name == animationName {
		return
	}
	if player.blendAnimation != nil && player.blendAnimation.Name == animationName {
		return
	}
	if player.secondaryAnimation != nil && animationName == *player.secondaryAnimation {
		return
	}

	if blendAnimation, ok := player.animations[animationName]; ok {
		player.blendAnimation = blendAnimation
		player.elapsedTime = 0
		player.blendAnimationElapsedTime = 0
		player.blendDuration = blendDuration
		player.blendDurationSoFar = 0
	} else {
		panic(fmt.Sprintf("failed to find animation %s", animationName))
	}
}

func (player *AnimationPlayer) PlayOnce(animationName string, secondaryAnimation string, blendDuration time.Duration) {
	local := secondaryAnimation
	player.secondaryAnimation = &local

	if blendAnimation, ok := player.animations[animationName]; ok {
		player.blendAnimation = blendAnimation
		player.elapsedTime = 0
		player.blendAnimationElapsedTime = 0
		player.blendDuration = blendDuration
		player.blendDurationSoFar = 0
		player.loop = false
	} else {
		panic(fmt.Sprintf("failed to find animation %s", animationName))
	}
}

func (player *AnimationPlayer) Update(delta time.Duration) {
	if player.currentAnimation == nil {
		return
	}

	player.elapsedTime += delta
	player.blendAnimationElapsedTime += delta
	player.blendDurationSoFar += delta

	for player.elapsedTime.Milliseconds() > player.currentAnimation.Length.Milliseconds() {
		player.elapsedTime = time.Duration(player.elapsedTime.Milliseconds()-player.currentAnimation.Length.Milliseconds()) * time.Millisecond

		// if we're not looping, we should have a secondary animation to fall back into
		if !player.loop {
			player.loop = true
			anim := player.secondaryAnimation
			player.secondaryAnimation = nil
			player.PlayAndBlendAnimation(*anim, 250*time.Millisecond)
		}
	}

	pose := player.calcPose(player.elapsedTime, player.currentAnimation)
	if player.blendAnimation != nil {
		for player.blendAnimationElapsedTime.Milliseconds() > player.blendAnimation.Length.Milliseconds() {
			player.blendAnimationElapsedTime = time.Duration(player.blendAnimationElapsedTime.Milliseconds()-player.blendAnimation.Length.Milliseconds()) * time.Millisecond
		}
		blendTargetPose := player.calcPose(player.blendAnimationElapsedTime, player.blendAnimation)
		blendProgression := float32(player.blendDurationSoFar.Milliseconds()) / float32(player.blendDuration.Milliseconds())
		if blendProgression >= 1 {
			player.currentAnimation = player.blendAnimation
			player.blendAnimation = nil
		}
		pose = interpolatePoses(pose, blendTargetPose, blendProgression)
	}

	animationTransforms := player.computeAnimationTransforms(pose)
	player.animationTransforms = animationTransforms
}

func (player *AnimationPlayer) calcPose(elapsedTime time.Duration, animation *modelspec.AnimationSpec) map[int]*modelspec.JointTransform {
	pose := calculateCurrentAnimationPose(elapsedTime, animation.KeyFrames)
	return pose
}

func (player *AnimationPlayer) computeAnimationTransforms(pose map[int]*modelspec.JointTransform) map[int]mgl32.Mat4 {
	poseTransforms := convertPoseToTransformMatrix(pose)
	animationTransforms := computeJointTransforms(player.rootJoint, poseTransforms)
	return animationTransforms
}

// applyPoseToJoints returns the set of transforms that move the joint from the bind pose to the given pose
func computeJointTransforms(joint *modelspec.JointSpec, pose map[int]mgl32.Mat4) map[int]mgl32.Mat4 {
	animationTransforms := map[int]mgl32.Mat4{}
	computeJointTransformsHelper(joint, mgl32.Ident4(), pose, animationTransforms)
	return animationTransforms
}

func computeJointTransformsHelper(joint *modelspec.JointSpec, parentTransform mgl32.Mat4, pose map[int]mgl32.Mat4, transforms map[int]mgl32.Mat4) {
	localTransform := pose[joint.ID]

	if _, ok := pose[joint.ID]; !ok {
		// panic(fmt.Sprintf("joint with id %d does not have a pose", joint.ID))
		localTransform = joint.BindTransform
	}

	// model-space transform that includes all the parental transforms
	// and the local transform, not meant to be used to transform any vertices
	// until we multiply it by the inverse bind transform
	poseTransform := parentTransform.Mul4(localTransform)

	for _, child := range joint.Children {
		computeJointTransformsHelper(child, poseTransform, pose, transforms)
	}

	// this is the model-space transform that can finally be used to transform
	// any vertices it influences
	transforms[joint.ID] = poseTransform.Mul4(joint.InverseBindTransform)
}

func calculateCurrentAnimationPose(elapsedTime time.Duration, keyFrames []*modelspec.KeyFrame) map[int]*modelspec.JointTransform {
	var startKeyFrame *modelspec.KeyFrame
	var endKeyFrame *modelspec.KeyFrame
	var progression float32

	// iterate backwards looking for the starting keyframe
	for i := len(keyFrames) - 1; i >= 0; i-- {
		keyFrame := keyFrames[i]
		if elapsedTime >= keyFrame.Start || i == 0 {
			startKeyFrame = keyFrame
			if i < len(keyFrames)-1 {
				endKeyFrame = keyFrames[i+1]
				progression = float32((elapsedTime - startKeyFrame.Start).Milliseconds()) / float32((endKeyFrame.Start - startKeyFrame.Start).Milliseconds())
			} else {
				// interpolate towards the first kf, assume looping animations
				endKeyFrame = keyFrames[0]
				progression = 0
			}
			break
		}
	}

	// progression = 0
	// startKeyFrame = keyFrames[0]
	return interpolatePoses(startKeyFrame.Pose, endKeyFrame.Pose, progression)
}

func convertPoseToTransformMatrix(pose map[int]*modelspec.JointTransform) map[int]mgl32.Mat4 {
	transformMatrices := map[int]mgl32.Mat4{}
	for jointID, transform := range pose {
		translation := transform.Translation
		rotation := transform.Rotation.Mat4()
		scale := transform.Scale
		transformMatrices[jointID] = mgl32.Translate3D(translation.X(), translation.Y(), translation.Z()).Mul4(rotation).Mul4(mgl32.Scale3D(scale.X(), scale.Y(), scale.Z()))
	}
	return transformMatrices
}

func interpolatePoses(j1, j2 map[int]*modelspec.JointTransform, progression float32) map[int]*modelspec.JointTransform {
	if progression > 1 {
		progression = 1
	}

	if progression < 0 {
		progression = 0
	}

	interpolatedPose := map[int]*modelspec.JointTransform{}
	for jointID := range j1 {
		k1JointTransform := j1[jointID]
		k2JointTransform := j2[jointID]

		// WTF - this lerp doesn't look right when interpolating keyframes???
		// rotationQuat := mgl32.QuatLerp(k1JointTransform.Rotation, k2JointTransform.Rotation, progression)
		rotation := libutils.QInterpolate(k1JointTransform.Rotation, k2JointTransform.Rotation, progression)

		translation := k1JointTransform.Translation.Add(k2JointTransform.Translation.Sub(k1JointTransform.Translation).Mul(progression))
		scale := k1JointTransform.Scale.Add(k2JointTransform.Scale.Sub(k1JointTransform.Scale).Mul(progression))

		// interpolatedPose[jointID] = mgl32.Translate3D(translation.X(), translation.Y(), translation.Z()).Mul4(rotation).Mul4(mgl32.Scale3D(scale.X(), scale.Y(), scale.Z()))
		interpolatedPose[jointID] = &modelspec.JointTransform{
			Translation: translation,
			Rotation:    rotation,
			Scale:       scale,
		}
	}
	return interpolatedPose
}
