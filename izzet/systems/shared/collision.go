package shared

import (
	"fmt"
	"sort"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/izzet/izzet/world"
	"github.com/kkevinchou/kitolib/collision"
	"github.com/kkevinchou/kitolib/collision/collider"
)

type App interface {
	CommandFrame() int
	IsClient() bool
	IsServer() bool
	World() *world.GameWorld
	GetPlayerEntity() *entities.Entity
}

type ICollisionObserver interface {
	OnBoundingBoxCheck(e1 *entities.Entity, e2 *entities.Entity)
	OnSpatialQuery(entityID int, count int)
	OnCollisionCheck(e1 *entities.Entity, e2 *entities.Entity)
	OnCollisionResolution(entityID int)
	Clear()
}

const (
	resolveCountMax   int     = 3
	GroundedThreshold float64 = 0.85
)

type packedIdxPair struct{ PackedIndexA, PackedIndexB int }

func pairKey(a, b int) packedIdxPair {
	if a < b {
		return packedIdxPair{PackedIndexA: a, PackedIndexB: b}
	}
	return packedIdxPair{PackedIndexA: b, PackedIndexB: a}
}

type collisionContext struct {
	world                *world.GameWorld
	observer             ICollisionObserver
	pairs                []packedIdxPair
	localPlayerCollision bool

	packedCollisionData []collisionData
	idToPackedIdx       map[int]int
	contacts            []collision.Contact
}

// no pointers, meant to be very quickly iterated
type collisionData struct {
	entityID               int
	shouldResolve          bool
	collisionMask          types.ColliderGroupFlag
	boundingBox            collider.BoundingBox
	boundingBoxInitialized bool

	capsuleCollider     collider.Capsule
	triMeshCollider     collider.TriMesh
	colliderInitialized bool

	hasCapsuleCollider bool
	hasTriMeshCollider bool

	static bool
}

func ResolveCollisions(app App, observer ICollisionObserver) {
	context := NewCollisionContext(app, observer)
	broadPhaseCollectPairs(context)
	detectAndResolve(context)
	postProcessing(context)
}

func NewCollisionContext(app App, observer ICollisionObserver) *collisionContext {
	context := &collisionContext{
		world:                app.World(),
		localPlayerCollision: app.IsClient(),
		observer:             observer,
		packedCollisionData:  nil,
		idToPackedIdx:        make(map[int]int),
	}

	// set up the full list of entities that can be involved in collisions
	for _, e := range app.World().Entities() {
		if e.Collider == nil {
			continue
		}

		context.idToPackedIdx[e.GetID()] = len(context.packedCollisionData)
		var cd collisionData
		cd.entityID = e.GetID()
		cd.collisionMask = e.Collider.CollisionMask
		cd.static = e.Static

		if context.localPlayerCollision {
			if e.GetID() == app.GetPlayerEntity().GetID() {
				cd.shouldResolve = true
			}
		} else {
			// while static entities can be involved in collisions
			// we do not need to resolve collisions for them since they do not move
			// we will rely on resolving the non-static entities that collide with them
			cd.shouldResolve = !e.Static
		}

		context.packedCollisionData = append(context.packedCollisionData, cd)
	}

	return context
}

func broadPhaseCollectPairs(context *collisionContext) {
	world := context.world
	observer := context.observer
	uniquePairMap := map[packedIdxPair]any{}

	for i, cd := range context.packedCollisionData {
		if !cd.shouldResolve {
			continue
		}
		setupTransformedBoundingBox(context, i)
		candidates := world.SpatialPartition().QueryEntities(context.packedCollisionData[i].boundingBox)
		observer.OnSpatialQuery(cd.entityID, len(candidates))
		for _, c := range candidates {
			e2 := world.GetEntityByID(c.GetID())
			// TODO: remove this hack, the nil check handles deleted entities
			if e2 == nil || e2.Collider == nil || cd.entityID == e2.ID {
				continue
			}

			if cd.collisionMask&e2.Collider.ColliderGroup == 0 {
				continue
			}

			key := pairKey(i, context.idToPackedIdx[e2.GetID()])
			if _, ok := uniquePairMap[key]; ok {
				continue
			}

			uniquePairMap[key] = struct{}{}
			context.pairs = append(context.pairs, key)
		}
	}
}

func detectAndResolve(context *collisionContext) {
	// 1. collect pairs of entities that are colliding, sorted by separating vector
	// 2. perform collision resolution for any colliding entities
	// 3. this can cause more collisions, repeat until no more further detected collisions, or we hit the configured max

	maxRunCount := len(context.pairs) * 2 * resolveCountMax // very rough estimate

	// collisionRuns acts as a fail safe for when we run collision detection and resolution infinitely.
	// given that we have a cap on collision resolution for each entity, we should never run more than
	// the number of entities times the cap.
	collisionRuns := 0
	for collisionRuns = 0; collisionRuns < maxRunCount; collisionRuns++ {
		collisionCandidates := collectSortedCollisionCandidates(context)
		if len(collisionCandidates) == 0 {
			break
		}

		contact := collisionCandidates[0]

		resolveCollision(context, contact, context.observer)

		// mark colliders and bounding boxes as uninitialized so future iterations will initialize them
		if !context.packedCollisionData[contact.PackedIndexA].static {
			context.packedCollisionData[contact.PackedIndexA].boundingBoxInitialized = false
			context.packedCollisionData[contact.PackedIndexA].colliderInitialized = false
		}
		if !context.packedCollisionData[contact.PackedIndexB].static {
			context.packedCollisionData[contact.PackedIndexB].boundingBoxInitialized = false
			context.packedCollisionData[contact.PackedIndexB].colliderInitialized = false
		}
		context.contacts = append(context.contacts, contact)
	}

	if collisionRuns == maxRunCount && collisionRuns != 0 {
		fmt.Println("hit absolute max collision run count", collisionRuns)
	}
}

func setupTransformedBoundingBox(context *collisionContext, packedIndex int) {
	if !context.packedCollisionData[packedIndex].boundingBoxInitialized {
		context.packedCollisionData[packedIndex].boundingBox = context.world.GetEntityByID(context.packedCollisionData[packedIndex].entityID).BoundingBox()
		context.packedCollisionData[packedIndex].boundingBoxInitialized = true
	}
}

func setupTransformedCollider(context *collisionContext, packedIndex int) {
	if !context.packedCollisionData[packedIndex].colliderInitialized {
		entity := context.world.GetEntityByID(context.packedCollisionData[packedIndex].entityID)
		cc := entity.Collider
		if cc.CapsuleCollider != nil {
			transformMatrix := entities.WorldTransform(entity)
			capsule := cc.CapsuleCollider.Transform(transformMatrix)
			context.packedCollisionData[packedIndex].capsuleCollider = capsule
			context.packedCollisionData[packedIndex].hasCapsuleCollider = true
		} else if cc.TriMeshCollider != nil {
			transformMatrix := entities.WorldTransform(entity)
			var triMesh collider.TriMesh
			if cc.SimplifiedTriMeshCollider != nil {
				triMesh = cc.SimplifiedTriMeshCollider.Transform(transformMatrix)
			} else {
				triMesh = cc.TriMeshCollider.Transform(transformMatrix)
			}
			context.packedCollisionData[packedIndex].triMeshCollider = triMesh
			context.packedCollisionData[packedIndex].hasTriMeshCollider = true
		}
	}
	context.packedCollisionData[packedIndex].colliderInitialized = true
}

func collectSortedCollisionCandidates(context *collisionContext) []collision.Contact {
	var allContacts []collision.Contact
	for _, pair := range context.pairs {
		setupTransformedBoundingBox(context, pair.PackedIndexA)
		setupTransformedBoundingBox(context, pair.PackedIndexB)

		if !collideBoundingBox(context.packedCollisionData[pair.PackedIndexA].boundingBox, context.packedCollisionData[pair.PackedIndexB].boundingBox) {
			continue
		}

		setupTransformedCollider(context, pair.PackedIndexA)
		setupTransformedCollider(context, pair.PackedIndexB)

		contacts := collide(context, pair.PackedIndexA, pair.PackedIndexB)

		if len(contacts) == 0 {
			continue
		}

		allContacts = append(allContacts, contacts...)
	}

	// TODO: optimization here would be to just maintain the shallowest collision rather than sorting
	sort.Sort(collision.ContactsBySeparatingDistance(allContacts))

	return allContacts
}

func collideBoundingBox(bb1, bb2 collider.BoundingBox) bool {
	// observer.OnBoundingBoxCheck(e1, e2)

	if bb1.MaxVertex.X() < bb2.MinVertex.X() || bb2.MaxVertex.X() < bb1.MinVertex.X() {
		return false
	}

	if bb1.MaxVertex.Y() < bb2.MinVertex.Y() || bb2.MaxVertex.Y() < bb1.MinVertex.Y() {
		return false
	}

	if bb1.MaxVertex.Z() < bb2.MinVertex.Z() || bb2.MaxVertex.Z() < bb1.MinVertex.Z() {
		return false
	}

	return true
}

func collide(context *collisionContext, a, b int) []collision.Contact {
	var result []collision.Contact

	collisionDataA := context.packedCollisionData[a]
	collisionDataB := context.packedCollisionData[b]

	if (collisionDataA.hasCapsuleCollider && collisionDataB.hasTriMeshCollider) || (collisionDataB.hasCapsuleCollider && collisionDataA.hasTriMeshCollider) {
		var capsuleCollider collider.Capsule
		var triMeshCollider collider.TriMesh

		if collisionDataA.hasCapsuleCollider {
			capsuleCollider = collisionDataA.capsuleCollider
			triMeshCollider = collisionDataB.triMeshCollider
		} else {
			capsuleCollider = collisionDataB.capsuleCollider
			triMeshCollider = collisionDataA.triMeshCollider
		}

		contacts := collision.CheckCollisionCapsuleTriMesh(
			capsuleCollider,
			triMeshCollider,
		)

		if len(contacts) == 0 {
			return nil
		}

		for _, contact := range contacts {
			c := collision.Contact{
				PackedIndexA:       a,
				PackedIndexB:       b,
				Type:               contact.Type,
				SeparatingVector:   contact.SeparatingVector,
				SeparatingDistance: contact.SeparatingDistance,
			}
			if collisionDataB.hasCapsuleCollider {
				c.SeparatingVector = c.SeparatingVector.Mul(-1)
			}
			result = append(result, c)
		}
	} else if collisionDataA.hasCapsuleCollider && collisionDataB.hasCapsuleCollider {
		contact, collisionDetected := collision.CheckCollisionCapsuleCapsule(
			collisionDataA.capsuleCollider,
			collisionDataB.capsuleCollider,
		)

		if !collisionDetected {
			return nil
		}

		result = append(result, collision.Contact{
			PackedIndexA:       a,
			PackedIndexB:       b,
			Type:               contact.Type,
			SeparatingVector:   contact.SeparatingVector,
			SeparatingDistance: contact.SeparatingDistance,
		})
	}

	// filter out contacts that have tiny separating distances
	threshold := 0.00005
	var filteredContacts []collision.Contact
	for _, contact := range result {
		if contact.SeparatingDistance > threshold {
			filteredContacts = append(filteredContacts, contact)
		}
	}
	return filteredContacts
}

func resolveCollision(context *collisionContext, contact collision.Contact, observer ICollisionObserver) {
	separatingVector := contact.SeparatingVector
	collisionDataA := context.packedCollisionData[contact.PackedIndexA]
	collisionDataB := context.packedCollisionData[contact.PackedIndexB]

	if context.localPlayerCollision {
		// allocate the whole separating distance to the player. the player is denoted with "shouldResolve"
		// maybe there's a cleaner way to do this?
		if collisionDataA.shouldResolve {
			entity := getEntity(context, contact.PackedIndexA)
			entities.SetLocalPosition(entity, entity.GetLocalPosition().Add(separatingVector))
		} else if collisionDataB.shouldResolve {
			entity := getEntity(context, contact.PackedIndexB)
			entities.SetLocalPosition(entity, entity.GetLocalPosition().Sub(separatingVector))
		}
	} else {
		// allocate half the separating distance between the two entities
		entityA := getEntity(context, contact.PackedIndexA)
		entityB := getEntity(context, contact.PackedIndexB)

		var aOnly bool
		var bOnly bool
		if entityA.CharacterControllerComponent != nil && entityA.CharacterControllerComponent.ControlVector.Len() > 0 {
			aOnly = true
		}
		if entityB.CharacterControllerComponent != nil && entityB.CharacterControllerComponent.ControlVector.Len() > 0 {
			bOnly = true
		}

		if !entityA.Static {
			var factor float64 = 0.5
			if entityB.Static {
				factor = 1
			}
			if aOnly {
				factor = 1
			} else if bOnly {
				factor = 0
			}
			entities.SetLocalPosition(entityA, entityA.GetLocalPosition().Add(separatingVector.Mul(factor)))
		}

		if !entityB.Static {
			var factor float64 = 0.5
			if entityA.Static {
				factor = 1
			}
			if bOnly {
				factor = 1
			} else if aOnly {
				factor = 0
			}
			entities.SetLocalPosition(entityB, entityB.GetLocalPosition().Sub(separatingVector.Mul(factor)))
		}
	}
}

func postProcessing(context *collisionContext) {
	for _, pair := range context.pairs {
		e1 := getEntity(context, pair.PackedIndexA)
		if e1.Physics != nil {
			e1.Physics.Grounded = false
		}
		e2 := getEntity(context, pair.PackedIndexB)
		if e2.Physics != nil {
			e2.Physics.Grounded = false
		}
	}

	for _, contact := range context.contacts {
		if contact.SeparatingVector.Normalize().Dot(mgl64.Vec3{0, 1, 0}) > GroundedThreshold {
			entity := getEntity(context, contact.PackedIndexA)
			entity.Kinematic.Grounded = true
			entity.Kinematic.Velocity = mgl64.Vec3{0, 0, 0}
		}
		if contact.SeparatingVector.Normalize().Dot(mgl64.Vec3{0, -1, 0}) > GroundedThreshold {
			entity := getEntity(context, contact.PackedIndexB)
			entity.Kinematic.Grounded = true
			entity.Kinematic.Velocity = mgl64.Vec3{0, 0, 0}
		}
	}
}

func getEntity(context *collisionContext, index int) *entities.Entity {
	return context.world.GetEntityByID(context.packedCollisionData[index].entityID)
}
