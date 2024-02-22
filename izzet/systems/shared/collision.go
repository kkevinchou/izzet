package shared

import (
	"fmt"
	"sort"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/app"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/kitolib/collision"
)

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

func ResolveCollisionsSingle(world GameWorld, entity *entities.Entity, observer ICollisionObserver) {
	ResolveCollisions(world, []*entities.Entity{entity}, observer)
}

func ResolveCollisions(world GameWorld, worldEntities []*entities.Entity, observer ICollisionObserver) {
	uniquePairMap := map[string]any{}
	seen := map[int]bool{}

	entityPairs := [][]*entities.Entity{}
	var entityList []*entities.Entity
	for _, e1 := range worldEntities {
		if e1.Static {
			// allow other entities to collide against a static entity. static entities
			// are picked up when the spatial partition query is done below for e2
			continue
		}
		if e1.Collider == nil {
			continue
		}

		entitiesInPartition := world.SpatialPartition().QueryEntities(e1.BoundingBox())
		observer.OnSpatialQuery(e1.GetID(), len(entitiesInPartition))
		for _, spatialEntity := range entitiesInPartition {
			e2 := world.GetEntityByID(spatialEntity.GetID())
			// todo: remove this hack, entities that are deleted should be removed
			// from the spatial partition
			if e2 == nil {
				continue
			}
			if e2.Collider == nil {
				continue
			}

			if e1.ID == e2.ID {
				continue
			}

			if e1.Collider.CollisionMask&e2.Collider.ColliderGroup == 0 {
				continue
			}

			if _, ok := uniquePairMap[generateUniquePairKey(e1, e2)]; ok {
				continue
			}
			if _, ok := uniquePairMap[generateUniquePairKey(e2, e1)]; ok {
				continue
			}

			if !seen[e1.ID] {
				entityList = append(entityList, e1)
			}
			if !seen[e2.ID] {
				entityList = append(entityList, e2)
			}

			entityPairs = append(entityPairs, []*entities.Entity{e1, e2})
			uniquePairMap[generateUniquePairKey(e1, e2)] = true
			uniquePairMap[generateUniquePairKey(e2, e1)] = true
		}
	}

	if len(entityPairs) > 0 {
		detectAndResolveCollisionsForEntityPairs(entityPairs, entityList, world, observer)
	}

	// reset contacts - TODO probably want to do this later in a separate system
	for _, entity := range worldEntities {
		if entity.Static || entity.Collider == nil {
			continue
		}

		if entity.Physics != nil {
			entity.Physics.Grounded = false
		}

		if entity.Collider.Contacts != nil && entity.Physics != nil {
			for _, contact := range entity.Collider.Contacts {
				if contact.SeparatingVector.Normalize().Dot(mgl64.Vec3{0, 1, 0}) > GroundedThreshold {
					entity.Physics.Grounded = true
					entity.Physics.Velocity = mgl64.Vec3{0, 0, 0}
				}
			}
		}

		entity.Collider.Contacts = nil
	}
}

func detectAndResolveCollisionsForEntityPairs(entityPairs [][]*entities.Entity, entityList []*entities.Entity, world GameWorld, observer ICollisionObserver) {
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
		// TODO: update entityPairs to not include collisions that have already been resolved
		// in fact, we may want to do the looping at the ResolveCollisions level
		collisionCandidates := collectSortedCollisionCandidates(entityPairs, entityList, maximallyCollidingEntities, world, observer)
		if len(collisionCandidates) == 0 {
			break
		}

		collisionCandidates = filterCollisionCandidates(collisionCandidates)

		for _, contact := range collisionCandidates {
			entity := world.GetEntityByID(*contact.EntityID)
			sourceEntity := world.GetEntityByID(*contact.SourceEntityID)
			resolveCollision(entity, sourceEntity, contact, observer)

			entity.Collider.Contacts = append(entity.Collider.Contacts, *contact)

			if resolveCount[entity.ID] > resolveCountMax {
				maximallyCollidingEntities[entity.ID] = true
			}
		}
	}

	if collisionRuns == absoluteMaxRunCount && collisionRuns != 0 {
		fmt.Println("hit absolute max collision run count", collisionRuns)
	}
}

// only resolve collisions for an entity, one collision at a time. resolving one collision may, as a side-effect, resolve other collisions as well
func filterCollisionCandidates(contacts []*collision.Contact) []*collision.Contact {
	var filteredSlice []*collision.Contact

	seen := map[int]bool{}
	for i := range contacts {
		contact := contacts[i]
		if _, ok := seen[*contact.EntityID]; !ok {
			seen[*contact.EntityID] = true
			filteredSlice = append(filteredSlice, contact)
		}
	}

	return filteredSlice
}

// collectSortedCollisionCandidates collects all potential collisions that can occur in the frame.
// these are "candidates" in that they are not guaranteed to have actually happened since
// if we resolve some of the collisions in the list, some will be invalidated
func collectSortedCollisionCandidates(entityPairs [][]*entities.Entity, entityList []*entities.Entity, skipEntitySet map[int]bool, world GameWorld, observer ICollisionObserver) []*collision.Contact {
	// initialize collision state

	for _, e := range entityList {
		cc := e.Collider
		if cc.CapsuleCollider != nil {
			transformMatrix := entities.WorldTransform(e)
			capsule := cc.CapsuleCollider.Transform(transformMatrix)
			cc.TransformedCapsuleCollider = &capsule
		} else if cc.TriMeshCollider != nil && (!e.Static || cc.TransformedTriMeshCollider == nil) {
			// localPosition := entities.GetLocalPosition(e)
			// transformMatrix := mgl64.Translate3D(localPosition.X(), localPosition.Y(), localPosition.Z())
			transformMatrix := entities.WorldTransform(e)
			// triMesh := cc.TriMeshCollider.Transform(transformMatrix)
			triMesh := cc.TriMeshCollider.Transform(transformMatrix)
			if cc.SimplifiedTriMeshCollider != nil {
				triMesh = cc.SimplifiedTriMeshCollider.Transform(transformMatrix)
			}
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

		if !collideBoundingBox(pair[0], pair[1], observer) {
			continue
		}

		contacts := collide(pair[0], pair[1], observer)
		if len(contacts) == 0 {
			continue
		}

		allContacts = append(allContacts, contacts...)
	}
	sort.Sort(collision.ContactsBySeparatingDistance(allContacts))

	return allContacts
}

func collideBoundingBox(e1 *entities.Entity, e2 *entities.Entity, observer ICollisionObserver) bool {
	observer.OnBoundingBoxCheck(e1, e2)

	bb1 := e1.BoundingBox()
	bb2 := e2.BoundingBox()

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

func collide(e1 *entities.Entity, e2 *entities.Entity, observer ICollisionObserver) []*collision.Contact {
	observer.OnCollisionCheck(e1, e2)

	var result []*collision.Contact

	if ok, capsuleEntity, triMeshEntity := app.IsCapsuleTriMeshCollision(e1, e2); ok {
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
	} else if ok := app.IsCapsuleCapsuleCollision(e1, e2); ok {
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

func resolveCollision(entity *entities.Entity, sourceEntity *entities.Entity, contact *collision.Contact, observer ICollisionObserver) {
	separatingVector := contact.SeparatingVector
	entities.SetLocalPosition(entity, entities.GetLocalPosition(entity).Add(separatingVector))
	observer.OnCollisionResolution(entity.GetID())
}

func generateUniquePairKey(e1, e2 *entities.Entity) string {
	return fmt.Sprintf("%d_%d", e1.GetID(), e2.GetID())
}
