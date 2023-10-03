package model

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/kkevinchou/kitolib/modelspec"
)

type RenderModel interface {
	JointMap() map[int]*modelspec.JointSpec
	RootJoint() *modelspec.JointSpec
	Name() string
}

type ModelConfig struct {
	MaxAnimationJointWeights int
}

type Model struct {
	name        string
	document    *modelspec.Document
	modelConfig *ModelConfig
	vertices    []modelspec.Vertex

	translation mgl32.Vec3
	rotation    mgl32.Quat
	scale       mgl32.Vec3
}

func (m *Model) Name() string {
	return m.name
}

func (m *Model) RootJoint() *modelspec.JointSpec {
	return m.document.RootJoint
}

func (m *Model) Animations() map[string]*modelspec.AnimationSpec {
	return m.document.Animations
}

func (m *Model) JointMap() map[int]*modelspec.JointSpec {
	return m.document.JointMap
}

func (m *Model) Vertices() []modelspec.Vertex {
	return m.vertices
}

func (m *Model) Translation() mgl32.Vec3 {
	return m.translation
}

func (m *Model) Rotation() mgl32.Quat {
	return m.rotation
}

func (m *Model) Scale() mgl32.Vec3 {
	return m.scale
}
