package entities

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl64"

	"github.com/kkevinchou/kitolib/collision/collider"
)

type Entity struct {
	ID                          int
	Name                        string
	Billboard                   bool
	Physics                     *PhysicsComponent
	Collider                    *ColliderComponent
	Particles                   *ParticleGenerator
	IsSocket                    bool
	LightInfo                   *LightInfo
	ImageInfo                   *ImageInfo
	ShapeData                   []*ShapeData
	Material                    *MaterialComponent
	Animation                   *AnimationComponent
	RenderBlend                 *RenderBlend
	CameraComponent             *CameraComponent
	Static                      bool
	ClientSidePredicted         bool
	SimplifiedTriMeshIterations int

	Deadge bool

	// dirty flag caching world transform
	DirtyTransformFlag   bool       `json:"-"`
	cachedWorldTransform mgl64.Mat4 // TODO: initialize to identity

	// each Entity has their own transforms and animation player
	LocalPosition mgl64.Vec3
	LocalRotation mgl64.Quat
	LocalScale    mgl64.Vec3

	MeshComponent       *MeshComponent
	InternalBoundingBox collider.BoundingBox

	CharacterControllerComponent *CharacterControllerComponent

	// relationships
	Parent   *Entity         `json:"-"`
	Children map[int]*Entity `json:"-"`

	PlayerInput *PlayerInputComponent
	AIComponent *AIComponent

	SpawnPointComponent *SpawnPoint
	PositionSync        *PositionSync
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

func (e *Entity) HasBoundingBox() bool {
	return e.InternalBoundingBox != collider.EmptyBoundingBox
}

func (e *Entity) BoundingBox() collider.BoundingBox {
	modelMatrix := WorldTransform(e)
	return e.InternalBoundingBox.Transform(modelMatrix)
}
