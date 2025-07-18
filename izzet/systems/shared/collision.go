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
	// initialEnts         map[int]*entities.Entity
	// only contains the initial entities that we wish to perform resolution form
	// on the client this is the local player entity, on the server this will contain all entities
	// initialEnts []*entities.Entity

	initialEntsAndCandidates []*entities.Entity
	resolvedRuns             int

	packedCollisionData []collisionData
	idToPackedIdx       map[int]int
	contacts            []collision.Contact2
}

// no pointers, meant to be very quickly iterated
type collisionData struct {
	entityID      int
	shouldResolve bool
	boundingBox   collider.BoundingBox
	collisionMask types.ColliderGroupFlag

	// transforms?
	capsuleCollider collider.Capsule
	triMeshCollider collider.TriMesh

	hasCapsuleCollider bool
	hasTriMeshCollider bool

	collidersInitialized bool
	resolveCount         int
	grounded             bool
}

// var logs bool = true

// only the entities that are passed in will have collision detection and resolution.
// collision detection is still checked against all other entities - if if not present
// in the original ents list
func ResolveCollisions2(app App, observer ICollisionObserver) {
	context := NewCollisionContext(app, observer)
	broadPhaseCollectPairs(context)
	// if !context.localPlayerCollision && logs {
	// 	fmt.Println("----------------- COLLISION START")
	// }
	detectAndResolve(context)
	// if !context.localPlayerCollision && logs {
	// 	fmt.Println("----------------- COLLISION END")
	// }
	postProcessing(context)
	// detectAndResolve(context)
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
		context.packedCollisionData[i].boundingBox = e.BoundingBox()
		context.packedCollisionData[i].collisionMask = e.Collider.CollisionMask

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

	for i, collisionData := range context.packedCollisionData {
		context.idToPackedIdx[collisionData.entityID] = i
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
		candidates := world.SpatialPartition().QueryEntities(e1.boundingBox)
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
	initializeColliders(context)

	// 1. collect pairs of entities that are colliding, sorted by separating vector
	// 2. perform collision resolution for any colliding entities
	// 3. this can cause more collisions, repeat until no more further detected collisions, or we hit the configured max

	maxRunCount := len(context.pairs) * 2 * resolveCountMax // very rough estimate

	// collisionRuns acts as a fail safe for when we run collision detection and resolution infinitely.
	// given that we have a cap on collision resolution for each entity, we should never run more than
	// the number of entities times the cap.
	collisionRuns := 0
	for collisionRuns = 0; collisionRuns < maxRunCount; collisionRuns++ {
		// TODO: update entityPairs to not include collisions that have already been resolved
		// in fact, we may want to do the looping at the ResolveCollisions level
		collisionCandidates := collectSortedCollisionCandidates2(context)
		if len(collisionCandidates) == 0 {
			break
		}

		contact := collisionCandidates[0]
		resolveCollision2(context, contact, context.observer)
		if contact.SeparatingVector.Normalize().Dot(mgl64.Vec3{0, 1, 0}) > GroundedThreshold {
			context.packedCollisionData[contact.PackedIndexA].grounded = true
		}
		if contact.SeparatingVector.Normalize().Dot(mgl64.Vec3{0, -1, 0}) > GroundedThreshold {
			context.packedCollisionData[contact.PackedIndexB].grounded = true
		}

		setupTransformedCollider(context, contact.PackedIndexA)
		setupTransformedCollider(context, contact.PackedIndexB)

		context.contacts = append(context.contacts, contact)
	}

	if collisionRuns == maxRunCount && collisionRuns != 0 {
		fmt.Println("hit absolute max collision run count", collisionRuns)
	}
}

func initializeColliders(context *collisionContext) {
	for _, pair := range context.pairs {
		collisionData := context.packedCollisionData[pair.PackedIndexA]
		if !collisionData.collidersInitialized {
			setupTransformedCollider(context, pair.PackedIndexA)
		}
		collisionData = context.packedCollisionData[pair.PackedIndexB]
		if !collisionData.collidersInitialized {
			setupTransformedCollider(context, pair.PackedIndexB)
		}
	}
}

func setupTransformedCollider(context *collisionContext, packedIndex int) {
	entity := context.world.GetEntityByID(context.packedCollisionData[packedIndex].entityID)
	// TODO: seems like the old code only transform the colliders for non static entities
	// for tri meshes. is that what we actually want? seems like a bug
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

func collectSortedCollisionCandidates2(context *collisionContext) []collision.Contact2 {
	var allContacts []collision.Contact2
	for _, pair := range context.pairs {
		if !collideBoundingBox2(context.packedCollisionData[pair.PackedIndexA].boundingBox, context.packedCollisionData[pair.PackedIndexB].boundingBox, context.observer) {
			continue
		}

		contacts := collide2(context, pair.PackedIndexA, pair.PackedIndexB, context.observer)
		if len(contacts) == 0 {
			continue
		}

		allContacts = append(allContacts, contacts...)
	}

	// TODO: optimization here would be to just maintain the shallowest collision rather than sorting
	sort.Sort(collision.ContactsBySeparatingDistance2(allContacts))

	return allContacts
}

func collideBoundingBox2(bb1, bb2 collider.BoundingBox, observer ICollisionObserver) bool {
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

func collide2(context *collisionContext, a, b int, observer ICollisionObserver) []collision.Contact2 {
	// observer.OnCollisionCheck(e1, e2)

	var result []collision.Contact2

	collisionDataA := context.packedCollisionData[a]
	collisionDataB := context.packedCollisionData[b]

	// if ok, capsuleEntity, triMeshEntity := physicsutils.IsCapsuleTriMeshCollision(e1, e2); ok {
	if (collisionDataA.hasCapsuleCollider && collisionDataB.hasTriMeshCollider) || (collisionDataB.hasCapsuleCollider && collisionDataA.hasTriMeshCollider) {
		// TODO: make non pointer version of the method

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
				// reverse
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

		// if !context.localPlayerCollision {
		// 	fmt.Println(context.world.GetEntityByID(context.packedCollisionData[a].entityID).Position(), contact.SeparatingDistance)
		// 	logs = false
		// }

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

func resolveCollision2(context *collisionContext, contact collision.Contact2, observer ICollisionObserver) {
	// TODO: try resolution scheme that doesn't doesn't equally effect both entities?
	separatingVector := contact.SeparatingVector
	collisionDataA := context.packedCollisionData[contact.PackedIndexA]
	collisionDataB := context.packedCollisionData[contact.PackedIndexB]

	if context.localPlayerCollision {
		// allocate the whole separating distance to the player. the player is denoted with "shouldResolve"
		// maybe a cleaner way to do this?
		if collisionDataA.shouldResolve {
			entityA := context.world.GetEntityByID(collisionDataA.entityID)
			entities.SetLocalPosition(entityA, entityA.GetLocalPosition().Add(separatingVector))

			// if context.localPlayerCollision {
			// 	if apputils.Vec3ApproxEqualThreshold(entityA.Position(), mgl64.Vec3{}, 5) {
			// 		fmt.Println("HI")
			// 	}
			// } else {
			// 	if apputils.Vec3ApproxEqualThreshold(entityA.Position(), mgl64.Vec3{}, 5) {
			// 		fmt.Println("HI")
			// 	}
			// }
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

	// observer.OnCollisionResolution(entity.GetID())
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
		// if contact.SeparatingVector.Normalize().Dot(mgl64.Vec3{0, 1, 0}) > GroundedThreshold {
		// }
		// if contact.SeparatingVector.Normalize().Dot(mgl64.Vec3{0, -1, 0}) > GroundedThreshold {
		// 	context.packedCollisionData[contact.PackedIndexB].grounded = true
		// }

		if context.packedCollisionData[contact.PackedIndexA].grounded {
			entity := context.world.GetEntityByID(context.packedCollisionData[contact.PackedIndexA].entityID)
			entity.Physics.Grounded = true
			entity.Physics.Velocity = mgl64.Vec3{0, 0, 0}
		}

		if context.packedCollisionData[contact.PackedIndexB].grounded {
			entity := context.world.GetEntityByID(context.packedCollisionData[contact.PackedIndexB].entityID)
			entity.Physics.Grounded = true
			entity.Physics.Velocity = mgl64.Vec3{0, 0, 0}
		}
	}
	// for _, entity := range ents {
	// 	if entity.Static || entity.Collider == nil {
	// 		continue
	// 	}

	// 	if entity.Physics != nil {
	// 		entity.Physics.Grounded = false
	// 	}

	// 	if entity.Collider.Contacts != nil && entity.Physics != nil {
	// 		for _, contact := range entity.Collider.Contacts {
	// 			if contact.SeparatingVector.Normalize().Dot(mgl64.Vec3{0, 1, 0}) > GroundedThreshold {
	// 				entity.Physics.Grounded = true
	// 				entity.Physics.Velocity = mgl64.Vec3{0, 0, 0}
	// 			}
	// 		}
	// 	}
	// }
}

func ResolveCollisionsSingle(app App, entity *entities.Entity, observer ICollisionObserver) {
	// ResolveCollisionsOriginal(app, []*entities.Entity{entity}, observer)
	ResolveCollisions2(app, observer)
}

func ResolveCollisions(app App, ents []*entities.Entity, observer ICollisionObserver) {
	// ResolveCollisionsOriginal(app, ents, observer)
	ResolveCollisions2(app, observer)
}

// func ResolveCollisionsOriginal(app App, ents []*entities.Entity, observer ICollisionObserver) {
// 	world := app.World()
// 	uniquePairMap := map[string]any{}
// 	seen := map[int]bool{}

// 	entityPairs := [][]*entities.Entity{}
// 	var entityList []*entities.Entity
// 	for _, e1 := range ents {
// 		if e1.Static {
// 			// allow other entities to collide against a static entity. static entities
// 			// are picked up when the spatial partition query is done below for e2
// 			continue
// 		}
// 		if e1.Collider == nil {
// 			continue
// 		}

// 		entitiesInPartition := world.SpatialPartition().QueryEntities(e1.BoundingBox())
// 		observer.OnSpatialQuery(e1.GetID(), len(entitiesInPartition))
// 		for _, spatialEntity := range entitiesInPartition {
// 			e2 := world.GetEntityByID(spatialEntity.GetID())
// 			// todo: remove this hack, entities that are deleted should be removed
// 			// from the spatial partition
// 			if e2 == nil {
// 				continue
// 			}
// 			if e2.Collider == nil {
// 				continue
// 			}

// 			if e1.ID == e2.ID {
// 				continue
// 			}

// 			if e1.Collider.CollisionMask&e2.Collider.ColliderGroup == 0 {
// 				continue
// 			}

// 			if _, ok := uniquePairMap[generateUniquePairKey(e1, e2)]; ok {
// 				continue
// 			}
// 			if _, ok := uniquePairMap[generateUniquePairKey(e2, e1)]; ok {
// 				continue
// 			}

// 			if !seen[e1.ID] {
// 				entityList = append(entityList, e1)
// 			}
// 			if !seen[e2.ID] {
// 				entityList = append(entityList, e2)
// 			}

// 			entityPairs = append(entityPairs, []*entities.Entity{e1, e2})
// 			uniquePairMap[generateUniquePairKey(e1, e2)] = true
// 			uniquePairMap[generateUniquePairKey(e2, e1)] = true
// 		}
// 	}

// 	if len(entityPairs) > 0 {
// 		detectAndResolveCollisionsForEntityPairs(entityPairs, entityList, world, observer)
// 	}

// 	for _, entity := range ents {
// 		if entity.Static || entity.Collider == nil {
// 			continue
// 		}

// 		if entity.Physics != nil {
// 			entity.Physics.Grounded = false
// 		}

// 		if entity.Collider.Contacts != nil && entity.Physics != nil {
// 			for _, contact := range entity.Collider.Contacts {
// 				if contact.SeparatingVector.Normalize().Dot(mgl64.Vec3{0, 1, 0}) > GroundedThreshold {
// 					entity.Physics.Grounded = true
// 					entity.Physics.Velocity = mgl64.Vec3{0, 0, 0}
// 				}
// 			}
// 		}
// 	}
// }

// func detectAndResolveCollisionsForEntityPairs(entityPairs [][]*entities.Entity, entityList []*entities.Entity, world GameWorld, observer ICollisionObserver) {
// 	// 1. collect pairs of entities that are colliding, sorted by separating vector
// 	// 2. perform collision resolution for any colliding entities
// 	// 3. this can cause more collisions, repeat until no more further detected collisions, or we hit the configured max

// 	resolveCount := map[int]int{}
// 	maximallyCollidingEntities := map[int]bool{}
// 	absoluteMaxRunCount := len(entityList) * resolveCountMax

// 	// collisionRuns acts as a fail safe for when we run collision detection and resolution infinitely.
// 	// given that we have a cap on collision resolution for each entity, we should never run more than
// 	// the number of entities times the cap.
// 	collisionRuns := 0
// 	for collisionRuns = 0; collisionRuns < absoluteMaxRunCount; collisionRuns++ {
// 		// TODO: update entityPairs to not include collisions that have already been resolved
// 		// in fact, we may want to do the looping at the ResolveCollisions level
// 		collisionCandidates := collectSortedCollisionCandidates(entityPairs, entityList, maximallyCollidingEntities, world, observer)
// 		if len(collisionCandidates) == 0 {
// 			break
// 		}

// 		collisionCandidates = filterCollisionCandidates(collisionCandidates)

// 		for _, contact := range collisionCandidates {
// 			entity := world.GetEntityByID(*contact.EntityID)
// 			sourceEntity := world.GetEntityByID(*contact.SourceEntityID)
// 			resolveCollision(entity, sourceEntity, contact, observer)

// 			entity.Collider.Contacts = append(entity.Collider.Contacts, *contact)

// 			resolveCount[entity.ID]++
// 			if resolveCount[entity.ID] > resolveCountMax {
// 				maximallyCollidingEntities[entity.ID] = true
// 			}
// 		}
// 	}

// 	if collisionRuns == absoluteMaxRunCount && collisionRuns != 0 {
// 		fmt.Println("hit absolute max collision run count", collisionRuns)
// 	}
// }

// // only resolve collisions for an entity, one collision at a time. resolving one collision may, as a side-effect, resolve other collisions as well
// func filterCollisionCandidates(contacts []*collision.Contact) []*collision.Contact {
// 	var filteredSlice []*collision.Contact

// 	seen := map[int]bool{}
// 	for i := range contacts {
// 		contact := contacts[i]
// 		if _, ok := seen[*contact.EntityID]; !ok {
// 			seen[*contact.EntityID] = true
// 			filteredSlice = append(filteredSlice, contact)
// 		}
// 	}

// 	return filteredSlice
// }

// // collectSortedCollisionCandidates collects all potential collisions that can occur in the frame.
// // these are "candidates" in that they are not guaranteed to have actually happened since
// // if we resolve some of the collisions in the list, some will be invalidated
// func collectSortedCollisionCandidates(entityPairs [][]*entities.Entity, entityList []*entities.Entity, skipEntitySet map[int]bool, world GameWorld, observer ICollisionObserver) []*collision.Contact {
// 	// initialize collision state

// 	for _, e := range entityList {
// 		cc := e.Collider
// 		if cc.CapsuleCollider != nil {
// 			transformMatrix := entities.WorldTransform(e)
// 			capsule := cc.CapsuleCollider.Transform(transformMatrix)
// 			cc.TransformedCapsuleCollider = &capsule
// 		} else if cc.TriMeshCollider != nil && (!e.Static || cc.TransformedTriMeshCollider == nil) {
// 			transformMatrix := entities.WorldTransform(e)
// 			triMesh := cc.TriMeshCollider.Transform(transformMatrix)
// 			if cc.SimplifiedTriMeshCollider != nil {
// 				triMesh = cc.SimplifiedTriMeshCollider.Transform(transformMatrix)
// 			}
// 			cc.TransformedTriMeshCollider = &triMesh
// 		}
// 	}

// 	var allContacts []*collision.Contact
// 	for _, pair := range entityPairs {
// 		if _, ok := skipEntitySet[pair[0].ID]; ok {
// 			continue
// 		}
// 		if _, ok := skipEntitySet[pair[1].ID]; ok {
// 			continue
// 		}

// 		if !collideBoundingBox(pair[0], pair[1], observer) {
// 			continue
// 		}

// 		contacts := collide(pair[0], pair[1], observer)
// 		if len(contacts) == 0 {
// 			continue
// 		}

// 		allContacts = append(allContacts, contacts...)
// 	}
// 	sort.Sort(collision.ContactsBySeparatingDistance(allContacts))

// 	return allContacts
// }

// func collideBoundingBox(e1 *entities.Entity, e2 *entities.Entity, observer ICollisionObserver) bool {
// 	observer.OnBoundingBoxCheck(e1, e2)

// 	bb1 := e1.BoundingBox()
// 	bb2 := e2.BoundingBox()

// 	if bb1.MaxVertex.X() < bb2.MinVertex.X() || bb2.MaxVertex.X() < bb1.MinVertex.X() {
// 		return false
// 	}

// 	if bb1.MaxVertex.Y() < bb2.MinVertex.Y() || bb2.MaxVertex.Y() < bb1.MinVertex.Y() {
// 		return false
// 	}

// 	if bb1.MaxVertex.Z() < bb2.MinVertex.Z() || bb2.MaxVertex.Z() < bb1.MinVertex.Z() {
// 		return false
// 	}

// 	return true
// }

// func collide(e1 *entities.Entity, e2 *entities.Entity, observer ICollisionObserver) []*collision.Contact {
// 	observer.OnCollisionCheck(e1, e2)

// 	var result []*collision.Contact

// 	if ok, capsuleEntity, triMeshEntity := physicsutils.IsCapsuleTriMeshCollision(e1, e2); ok {
// 		contacts := collision.CheckCollisionCapsuleTriMesh(
// 			*capsuleEntity.Collider.TransformedCapsuleCollider,
// 			*triMeshEntity.Collider.TransformedTriMeshCollider,
// 		)
// 		if len(contacts) == 0 {
// 			return nil
// 		}

// 		triEntityID := triMeshEntity.ID
// 		capsuleEntityID := capsuleEntity.ID

// 		for _, contact := range contacts {
// 			contact.EntityID = &capsuleEntityID
// 			contact.SourceEntityID = &triEntityID
// 		}

// 		result = contacts
// 	} else if ok := physicsutils.IsCapsuleCapsuleCollision(e1, e2); ok {
// 		contact := collision.CheckCollisionCapsuleCapsule(
// 			*e1.Collider.TransformedCapsuleCollider,
// 			*e2.Collider.TransformedCapsuleCollider,
// 		)
// 		if contact == nil {
// 			return nil
// 		}

// 		e1ID := e1.ID
// 		e2ID := e2.ID
// 		contact.EntityID = &e1ID
// 		contact.SourceEntityID = &e2ID
// 		result = append(result, contact)
// 	}

// 	// filter out contacts that have tiny separating distances
// 	threshold := 0.00005
// 	var filteredContacts []*collision.Contact
// 	for _, contact := range result {
// 		if contact.SeparatingDistance > threshold {
// 			filteredContacts = append(filteredContacts, contact)
// 		}
// 	}
// 	return filteredContacts
// }

// func resolveCollision(entity *entities.Entity, sourceEntity *entities.Entity, contact *collision.Contact, observer ICollisionObserver) {
// 	separatingVector := contact.SeparatingVector
// 	if entity.CharacterControllerComponent != nil && sourceEntity.CharacterControllerComponent != nil {
// 		// todo: try resolution scheme that doesn't doesn't equally effect both entities?
// 		fmt.Println("SOURCE:", sourceEntity.GetID(), "ENTITY:", entity.GetID())
// 	}
// 	entities.SetLocalPosition(entity, entity.GetLocalPosition().Add(separatingVector))
// 	observer.OnCollisionResolution(entity.GetID())
// }

// func generateUniquePairKey(e1, e2 *entities.Entity) string {
// 	return fmt.Sprintf("%d_%d", e1.GetID(), e2.GetID())
// }
