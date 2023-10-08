package izzet

import (
	"fmt"
	"sort"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/kitolib/collision"
	"github.com/kkevinchou/kitolib/spatialpartition"
)

var resolveCountMax = 3

type World interface {
	GetEntityByID(id int) *entities.Entity
	Entities() []*entities.Entity
	SpatialPartition() *spatialpartition.SpatialPartition
}

func (g *Izzet) physicsStep(delta time.Duration) {
	allEntities := g.Entities()

	for _, entity := range allEntities {
		physicsComponent := entity.Physics
		if physicsComponent == nil {
			continue
		}

		entities.SetLocalPosition(entity, entities.GetLocalPosition(entity).Add(physicsComponent.Velocity.Mul(delta.Seconds())))
	}

	// ResolveCollisions(g)

	// reset contacts - probably want to do this later
	for _, entity := range allEntities {
		if entity.Collider == nil {
			continue
		}
		entity.Collider.Contacts = map[int]bool{}
	}
}

func ResolveCollisions(world World) {
	var entityList []*entities.Entity
	for _, e := range world.Entities() {
		if e.Collider == nil {
			continue
		}
		entityList = append(entityList, e)
	}

	// pairExists stores the pairs of entities that we've already created,
	// don't create a pair for both (e1, e2) and (e2, e1), just one of them
	pairExists := map[int]map[int]bool{}
	for _, e := range entityList {
		pairExists[e.ID] = map[int]bool{}
	}

	entityPairs := [][]*entities.Entity{}
	for _, e1 := range entityList {
		entitiesInPartition := world.SpatialPartition().QueryEntities(*e1.BoundingBox())
		for _, spatialEntity := range entitiesInPartition {
			e2 := world.GetEntityByID(spatialEntity.GetID())
			if e1.ID == e2.ID {
				continue
			}

			if e1.Collider.CollisionMask&e2.Collider.ColliderGroup == 0 {
				continue
			}

			if pairExists[e1.ID][e2.ID] || pairExists[e2.ID][e1.ID] {
				continue
			}

			entityPairs = append(entityPairs, []*entities.Entity{e1, e2})
			pairExists[e1.ID][e2.ID] = true
			pairExists[e2.ID][e1.ID] = true
		}
	}
	// entityList = nil
	detectAndResolveCollisionsForEntityPairs(entityPairs, entityList, world)
}

func detectAndResolveCollisionsForEntityPairs(entityPairs [][]*entities.Entity, entityList []*entities.Entity, world World) {
	// 1. collect pairs of entities that are colliding, sorted by separating vector
	// 2. perform collision resolution for any colliding entities
	// 3. this can cause more collisions, repeat until no more further detected collisions, or we hit the configured max

	resolveCount := map[int]int{}
	maximallyCollidingEntities := map[int]bool{}
	absoluteMaxRunCount := len(entityList) * resolveCountMax

	// collisionRuns acts as a fail safe for when we run collision detection and resolution infinitely.
	// given that we have a cap on collision resolution for each entity, we should never run more than
	// the number of entities times the cap.
	collisionRuns := 0
	for collisionRuns = 0; collisionRuns < absoluteMaxRunCount; collisionRuns++ {
		collisionCandidates := collectSortedCollisionCandidates(entityPairs, entityList, maximallyCollidingEntities, world)
		if len(collisionCandidates) == 0 {
			break
		}

		resolvedEntities := resolveCollisions(collisionCandidates, world)
		for entityID, otherEntityID := range resolvedEntities {
			e1 := world.GetEntityByID(entityID)
			e2 := world.GetEntityByID(otherEntityID)
			resolveCount[entityID] += 1

			if resolveCount[entityID] > resolveCountMax {
				maximallyCollidingEntities[entityID] = true
				// fmt.Println("reached max count for entity", entityID, e1.GetName(), "most recent collision with", otherEntityID, e2.GetName())
			}

			// TODO(kchou): consider that two of the same entity may collide twice
			// also, we may want to support colliding with individual mesh chunks of an
			// entity rather than consideration of the whole entity itself

			// NOTE(kchou): ideally we'd include the full collision information (contact point, separting vector, etc)
			// but I don't yet have a good story around registering this information for both entities. i.e. if A is colliding
			// with B, is B colliding with A? do we share half the separation between the two?
			e1.Collider.Contacts[e2.ID] = true
			e2.Collider.Contacts[e1.ID] = true
		}
	}

	if collisionRuns == absoluteMaxRunCount && collisionRuns != 0 {
		fmt.Println("hit absolute max collision run count", collisionRuns)
	}
}

// collectSortedCollisionCandidates collects all potential collisions that can occur in the frame.
// these are "candidates" in that they are not guaranteed to have actually happened since
// if we resolve some of the collisions in the list, some will be invalidated
func collectSortedCollisionCandidates(entityPairs [][]*entities.Entity, entityList []*entities.Entity, skipEntitySet map[int]bool, world World) []*collision.Contact {
	// initialize collision state

	// TODO: may not need to transform the collider since colliders will be children of the actual entity
	for _, e := range entityList {
		cc := e.Collider
		if cc.CapsuleCollider != nil {
			capsule := cc.CapsuleCollider.Transform(entities.GetLocalPosition(e))
			cc.TransformedCapsuleCollider = &capsule
		} else if cc.TriMeshCollider != nil {
			localPosition := entities.GetLocalPosition(e)
			transformMatrix := mgl64.Translate3D(localPosition.X(), localPosition.Y(), localPosition.Z())
			triMesh := cc.TriMeshCollider.Transform(transformMatrix)
			cc.TransformedTriMeshCollider = &triMesh
		}
	}

	var allContacts []*collision.Contact
	for _, pair := range entityPairs {
		if _, ok := skipEntitySet[pair[0].ID]; ok {
			continue
		}
		if _, ok := skipEntitySet[pair[1].ID]; ok {
			continue
		}

		contacts := collide(pair[0], pair[1])
		if len(contacts) == 0 {
			continue
		}

		allContacts = append(allContacts, contacts...)
	}
	sort.Sort(collision.ContactsBySeparatingDistance(allContacts))

	return allContacts
}

func collide(e1 *entities.Entity, e2 *entities.Entity) []*collision.Contact {
	var result []*collision.Contact

	if ok, capsuleEntity, triMeshEntity := isCapsuleTriMeshCollision(e1, e2); ok {
		contacts := collision.CheckCollisionCapsuleTriMesh(
			*capsuleEntity.Collider.TransformedCapsuleCollider,
			*triMeshEntity.Collider.TransformedTriMeshCollider,
		)
		if len(contacts) == 0 {
			return nil
		}

		triEntityID := triMeshEntity.ID
		capsuleEntityID := capsuleEntity.ID

		for _, contact := range contacts {
			contact.EntityID = &capsuleEntityID
			contact.SourceEntityID = &triEntityID
		}

		result = contacts
	} else if ok := isCapsuleCapsuleCollision(e1, e2); ok {
		contact := collision.CheckCollisionCapsuleCapsule(
			*e1.Collider.TransformedCapsuleCollider,
			*e2.Collider.TransformedCapsuleCollider,
		)
		if contact == nil {
			return nil
		}

		e1ID := e1.ID
		e2ID := e2.ID
		contact.EntityID = &e1ID
		contact.SourceEntityID = &e2ID
		result = append(result, contact)
	}

	// filter out contacts that have tiny separating distances
	threshold := 0.00005
	var filteredContacts []*collision.Contact
	for _, contact := range result {
		if contact.SeparatingDistance > threshold {
			filteredContacts = append(filteredContacts, contact)
		}
	}
	return filteredContacts
}

func resolveCollisions(contacts []*collision.Contact, world World) map[int]int {
	resolved := map[int]int{}
	for _, contact := range contacts {
		entity := world.GetEntityByID(*contact.EntityID)
		sourceEntity := world.GetEntityByID(*contact.SourceEntityID)
		resolveCollision(entity, sourceEntity, contact)

		resolved[*contact.EntityID] = *contact.SourceEntityID
		resolved[*contact.SourceEntityID] = *contact.EntityID
	}

	return resolved
}

func resolveCollision(entity *entities.Entity, sourceEntity *entities.Entity, contact *collision.Contact) {
	separatingVector := contact.SeparatingVector
	entities.SetLocalPosition(entity, entities.GetLocalPosition(entity).Add(separatingVector))
}

func isCapsuleTriMeshCollision(e1, e2 *entities.Entity) (bool, *entities.Entity, *entities.Entity) {
	if e1.Collider.CapsuleCollider != nil {
		if e2.Collider.TriMeshCollider != nil {
			return true, e1, e2
		}
	}

	if e2.Collider.CapsuleCollider != nil {
		if e1.Collider.TriMeshCollider != nil {
			return true, e2, e1
		}
	}

	return false, nil, nil
}

func isCapsuleCapsuleCollision(e1, e2 *entities.Entity) bool {
	if e1.Collider.CapsuleCollider != nil {
		if e2.Collider.CapsuleCollider != nil {
			return true
		}
	}

	return false
}
