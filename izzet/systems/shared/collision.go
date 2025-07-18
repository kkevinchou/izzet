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
	maxResolves          int
	pairs                []packedIdxPair
	localPlayerCollision bool

	packedCollisionData []collisionData
	idToPackedIdx       map[int]int
	contacts            []collision.Contact2
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

	collidersInitialized bool
	static               bool
}

func ResolveCollisions2(app App, observer ICollisionObserver) {
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
		maxResolves:          resolveCountMax,
		packedCollisionData:  make([]collisionData, len(app.World().Entities())),
		idToPackedIdx:        make(map[int]int, len(app.World().Entities())),
	}

	// set up the full list of entities that can be involved in collisions
	for i, e := range app.World().Entities() {
		if e.Collider == nil {
			continue
		}

		context.packedCollisionData[i].entityID = e.GetID()
		context.packedCollisionData[i].collisionMask = e.Collider.CollisionMask
		context.packedCollisionData[i].static = e.Static
		context.idToPackedIdx[e.GetID()] = i

		if context.localPlayerCollision {
			if e.GetID() == app.GetPlayerEntity().GetID() {
				context.packedCollisionData[i].shouldResolve = true
			}
		} else {
			// while static entities can be involved in collisions
			// we do not need to resolve collisions for them since they do not move
			// we will rely on resolving the non-static entities that collide with them
			context.packedCollisionData[i].shouldResolve = !e.Static
		}
	}

	return context
}

func broadPhaseCollectPairs(context *collisionContext) {
	world := context.world
	observer := context.observer
	uniquePairMap := map[packedIdxPair]any{}

	for i, e1 := range context.packedCollisionData {
		if !e1.shouldResolve {
			continue
		}
		bb := context.world.GetEntityByID(e1.entityID).BoundingBox()
		candidates := world.SpatialPartition().QueryEntities(bb)
		observer.OnSpatialQuery(e1.entityID, len(candidates))
		for _, c := range candidates {
			e2 := world.GetEntityByID(c.GetID())
			// TODO: remove this hack, the nil check handles deleted entities
			if e2 == nil || e2.Collider == nil || e1.entityID == e2.ID {
				continue
			}

			if e1.collisionMask&e2.Collider.ColliderGroup == 0 {
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
	// initializeColliders(context)

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

func collectSortedCollisionCandidates(context *collisionContext) []collision.Contact2 {
	var allContacts []collision.Contact2
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
	sort.Sort(collision.ContactsBySeparatingDistance2(allContacts))

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

func collide(context *collisionContext, a, b int) []collision.Contact2 {
	var result []collision.Contact2

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
			c := collision.Contact2{
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
		contact := collision.CheckCollisionCapsuleCapsule(
			collisionDataA.capsuleCollider,
			collisionDataB.capsuleCollider,
		)
		if contact == nil {
			return nil
		}

		result = append(result, collision.Contact2{
			PackedIndexA:       a,
			PackedIndexB:       b,
			Type:               contact.Type,
			SeparatingVector:   contact.SeparatingVector,
			SeparatingDistance: contact.SeparatingDistance,
		})
	}

	// filter out contacts that have tiny separating distances
	threshold := 0.00005
	var filteredContacts []collision.Contact2
	for _, contact := range result {
		if contact.SeparatingDistance > threshold {
			filteredContacts = append(filteredContacts, contact)
		}
	}
	return filteredContacts
}

func resolveCollision(context *collisionContext, contact collision.Contact2, observer ICollisionObserver) {
	separatingVector := contact.SeparatingVector
	collisionDataA := context.packedCollisionData[contact.PackedIndexA]
	collisionDataB := context.packedCollisionData[contact.PackedIndexB]

	if context.localPlayerCollision {
		// allocate the whole separating distance to the player. the player is denoted with "shouldResolve"
		// maybe there's a cleaner way to do this?
		if collisionDataA.shouldResolve {
			entityA := context.world.GetEntityByID(collisionDataA.entityID)
			entities.SetLocalPosition(entityA, entityA.GetLocalPosition().Add(separatingVector))
		} else if collisionDataB.shouldResolve {
			entityB := context.world.GetEntityByID(collisionDataB.entityID)
			entities.SetLocalPosition(entityB, entityB.GetLocalPosition().Sub(separatingVector))
		}
	} else {
		// allocate half the separating distance between the two entities
		entityA := context.world.GetEntityByID(collisionDataA.entityID)
		entityB := context.world.GetEntityByID(collisionDataB.entityID)

		if !entityA.Static {
			var factor float64 = 0.5
			if entityB.Static {
				factor = 1
			}
			entities.SetLocalPosition(entityA, entityA.GetLocalPosition().Add(separatingVector.Mul(factor)))
		}

		if !entityB.Static {
			var factor float64 = 0.5
			if entityA.Static {
				factor = 1
			}
			entities.SetLocalPosition(entityB, entityB.GetLocalPosition().Sub(separatingVector.Mul(factor)))
		}
	}
}

func postProcessing(context *collisionContext) {
	for _, pair := range context.pairs {
		e1 := context.world.GetEntityByID(context.packedCollisionData[pair.PackedIndexA].entityID)
		if e1.Physics != nil {
			e1.Physics.Grounded = false
		}
		e2 := context.world.GetEntityByID(context.packedCollisionData[pair.PackedIndexB].entityID)
		if e2.Physics != nil {
			e2.Physics.Grounded = false
		}
	}

	for _, contact := range context.contacts {
		if contact.SeparatingVector.Normalize().Dot(mgl64.Vec3{0, 1, 0}) > GroundedThreshold {
			entity := context.world.GetEntityByID(context.packedCollisionData[contact.PackedIndexA].entityID)
			entity.Physics.Grounded = true
			entity.Physics.Velocity = mgl64.Vec3{0, 0, 0}
		}
		if contact.SeparatingVector.Normalize().Dot(mgl64.Vec3{0, -1, 0}) > GroundedThreshold {
			entity := context.world.GetEntityByID(context.packedCollisionData[contact.PackedIndexB].entityID)
			entity.Physics.Grounded = true
			entity.Physics.Velocity = mgl64.Vec3{0, 0, 0}
		}
	}
}

func ResolveCollisionsSingle(app App, entity *entities.Entity, observer ICollisionObserver) {
	ResolveCollisions2(app, observer)
}

func ResolveCollisions(app App, ents []*entities.Entity, observer ICollisionObserver) {
	ResolveCollisions2(app, observer)
}
