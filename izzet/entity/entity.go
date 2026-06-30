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

	LocalPosition mgl64.Vec3
	LocalRotation mgl64.Quat
	LocalScale    mgl64.Vec3

	// Optional Components
	Physics         *PhysicsComponent
	Kinematic       *KinematicComponent
	Collider        *ColliderComponent
	Particles       *ParticleGenerator
	LightInfo       *LightInfo
	ImageComponent  *ImageComponent
	ShapeData       []*ShapeData
	Animation       *AnimationComponent
	RenderBlend     *RenderBlend
	CameraComponent *CameraComponent
	HealthComponent *HealthComponent

	// dirty flag caching world transform
	DirtyTransformFlag   bool       `json:"-"`
	cachedWorldTransform mgl64.Mat4 // TODO: initialize to identity

	MeshComponent *MeshComponent

	CharacterControllerComponent *CharacterControllerComponent

	// relationships
	Parent   *Entity         `json:"-"`
	Children map[int]*Entity `json:"-"`

	PlayerInput *PlayerInputComponent
	AIComponent *AIComponent

	NavigationComponent *NavigationComponent
	AttackComponent     *AttackComponent

	SpawnPointComponent    *SpawnPoint
	AimDownSightsComponent *AimDownSightsComponent
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
