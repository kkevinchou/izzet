package entities

import "github.com/kkevinchou/kitolib/collision/collider"

type ColliderGroup string

var (
	ColliderGroupTerrain ColliderGroup = "TERRAIN"
	ColliderGroupPlayer  ColliderGroup = "PLAYER"
)

var ColliderGroupMap map[ColliderGroup]ColliderGroupFlag = map[ColliderGroup]ColliderGroupFlag{
	ColliderGroupTerrain: ColliderGroupFlagTerrain,
	ColliderGroupPlayer:  ColliderGroupFlagPlayer,
}

var ColliderFlagToGroupName map[ColliderGroupFlag]ColliderGroup

func init() {
	ColliderFlagToGroupName = map[ColliderGroupFlag]ColliderGroup{}
	for k, v := range ColliderGroupMap {
		ColliderFlagToGroupName[v] = k
	}
}

type ColliderGroupFlag uint64

const (
	ColliderGroupFlagTerrain ColliderGroupFlag = 1 << 0
	ColliderGroupFlagPlayer  ColliderGroupFlag = 2 << 0
)

type ColliderComponent struct {
	// entities with the same collider group do not collide with each other
	ColliderGroup ColliderGroupFlag
	CollisionMask ColliderGroupFlag

	// Skip separation tells the collision system to skip the step of separating colliding entities
	// for the entity that owns this component
	SkipSeparation bool

	// Contacts marks which entities it collided with in the current frame
	Contacts map[int]bool

	CapsuleCollider     *collider.Capsule
	TriMeshCollider     *collider.TriMesh `json:"-"`
	BoundingBoxCollider *collider.BoundingBox

	// stores the transformed collider (e.g. if the entity moves)
	TransformedCapsuleCollider     *collider.Capsule     `json:"-"`
	TransformedTriMeshCollider     *collider.TriMesh     `json:"-"`
	TransformedBoundingBoxCollider *collider.BoundingBox `json:"-"`
}
