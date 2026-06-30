package entity

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl64"
)

type Entity struct {
	ID   int
	Name string

	Static bool
	Deadge bool

	LocalPosition mgl64.Vec3 `json:",omitempty"`
	LocalRotation mgl64.Quat `json:",omitempty"`
	LocalScale    mgl64.Vec3 `json:",omitempty"`

	// Optional Components
	Physics         *PhysicsComponent   `json:",omitempty"`
	Kinematic       *KinematicComponent `json:",omitempty"`
	Collider        *ColliderComponent  `json:",omitempty"`
	Particles       *ParticleGenerator  `json:",omitempty"`
	LightInfo       *LightInfo          `json:",omitempty"`
	ImageComponent  *ImageComponent     `json:",omitempty"`
	ShapeData       []*ShapeData        `json:",omitempty"`
	Animation       *AnimationComponent `json:",omitempty"`
	RenderBlend     *RenderBlend        `json:",omitempty"`
	CameraComponent *CameraComponent    `json:",omitempty"`
	HealthComponent *HealthComponent    `json:",omitempty"`

	// dirty flag caching world transform
	DirtyTransformFlag   bool       `json:"-"`
	cachedWorldTransform mgl64.Mat4 // TODO: initialize to identity

	MeshComponent *MeshComponent `json:",omitempty"`

	CharacterControllerComponent *CharacterControllerComponent `json:",omitempty"`

	// relationships
	Parent   *Entity         `json:"-"`
	Children map[int]*Entity `json:"-"`

	PlayerInput *PlayerInputComponent `json:",omitempty"`
	AIComponent *AIComponent          `json:",omitempty"`

	NavigationComponent *NavigationComponent `json:",omitempty"`
	AttackComponent     *AttackComponent     `json:",omitempty"`

	SpawnPointComponent    *SpawnPoint             `json:",omitempty"`
	AimDownSightsComponent *AimDownSightsComponent `json:",omitempty"`
}

func (e *Entity) GetID() int {
	return e.ID
}

func (e *Entity) GetName() string {
	return e.Name
}

func (e *Entity) Dirty() bool {
	return e.DirtyTransformFlag
}

func (e *Entity) NameID() string {
	return fmt.Sprintf("%s-%d", e.Name, e.ID)
}

const (
	InvalidEntityID int = -1
)
