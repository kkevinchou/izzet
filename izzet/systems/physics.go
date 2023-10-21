package systems

import (
	"fmt"
	"sort"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/kitolib/collision"
)

const (
	resolveCountMax          int     = 3
	groundedThreshold        float64 = 0.85
	accelerationDueToGravity float64 = 250 // units per second
)

type PhysicsSystem struct {
}

func (s *PhysicsSystem) Update(delta time.Duration, world GameWorld) {
	allEntities := world.Entities()

	for _, entity := range allEntities {
		physicsComponent := entity.Physics
		if entity.Static || physicsComponent == nil {
			continue
		}

		if physicsComponent.GravityEnabled {
			velocityFromGravity := mgl64.Vec3{0, -accelerationDueToGravity * float64(delta.Milliseconds()) / 1000}
			physicsComponent.Velocity = physicsComponent.Velocity.Add(velocityFromGravity)
		}
		entities.SetLocalPosition(entity, entities.GetLocalPosition(entity).Add(physicsComponent.Velocity.Mul(delta.Seconds())))
	}

	ResolveCollisions(world)

	// reset contacts - probably want to do this later
	for _, entity := range allEntities {
		if entity.Collider == nil {
			continue
		}

		if entity.Physics != nil {
			entity.Physics.Grounded = false
		}

		if entity.Collider.Contacts != nil && entity.Physics != nil {
			for _, contact := range entity.Collider.Contacts {
				if contact.SeparatingVector.Normalize().Dot(mgl64.Vec3{0, 1, 0}) > groundedThreshold {
					entity.Physics.Grounded = true
					entity.Physics.Velocity = mgl64.Vec3{0, 0, 0}
				}
			}
		}

		// entity.Collider.Contacts = map[int]bool{}
		entity.Collider.Contacts = nil
	}
}

func ResolveCollisions(world GameWorld) {
	var collidableEntities []*entities.Entity
	for _, e := range world.Entities() {
		if e.Collider == nil {
			continue
		}
		collidableEntities = append(collidableEntities, e)
	}

	// pairExists stores the pairs of entities that we've already created,
	// don't create a pair for both (e1, e2) and (e2, e1), just one of them
	pairExists := map[int]map[int]bool{}
	for _, e := range collidableEntities {
		pairExists[e.ID] = map[int]bool{}
	}

	seen := map[int]bool{}

	entityPairs := [][]*entities.Entity{}
	var entityList []*entities.Entity
	for _, e1 := range collidableEntities {
		if e1.Static {
			// allow other entities to collide against a static entity. static entities
			// are picked up when the spatial partition query is done below for e2
			continue
		}

		entitiesInPartition := world.SpatialPartition().QueryEntities(e1.BoundingBox())
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

			if pairExists[e1.ID][e2.ID] || pairExists[e2.ID][e1.ID] {
				continue
			}

			if !seen[e1.ID] {
				entityList = append(entityList, e1)
			}
			if !seen[e2.ID] {
				entityList = append(entityList, e2)
			}

			entityPairs = append(entityPairs, []*entities.Entity{e1, e2})
			pairExists[e1.ID][e2.ID] = true
			pairExists[e2.ID][e1.ID] = true
		}
	}

	if len(entityPairs) > 0 {
		detectAndResolveCollisionsForEntityPairs(entityPairs, entityList, world)
	}
}

func detectAndResolveCollisionsForEntityPairs(entityPairs [][]*entities.Entity, entityList []*entities.Entity, world GameWorld) {
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
		collisionCandidates := collectSortedCollisionCandidates(entityPairs, entityList, maximallyCollidingEntities, world)
		if len(collisionCandidates) == 0 {
			break
		}

		collisionCandidates = filterCollisionCandidates(collisionCandidates)

		for _, contact := range collisionCandidates {
			entity := world.GetEntityByID(*contact.EntityID)
			sourceEntity := world.GetEntityByID(*contact.SourceEntityID)
			resolveCollision(entity, sourceEntity, contact)

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
func collectSortedCollisionCandidates(entityPairs [][]*entities.Entity, entityList []*entities.Entity, skipEntitySet map[int]bool, world GameWorld) []*collision.Contact {
	// initialize collision state

	// TODO: may not need to transform the collider since colliders will be children of the actual entity
	for _, e := range entityList {
		cc := e.Collider
		if cc.CapsuleCollider != nil {
			transformMatrix := entities.WorldTransform(e)
			capsule := cc.CapsuleCollider.Transform(transformMatrix)
			cc.TransformedCapsuleCollider = &capsule
		} else if cc.TriMeshCollider != nil {
			// localPosition := entities.GetLocalPosition(e)
			// transformMatrix := mgl64.Translate3D(localPosition.X(), localPosition.Y(), localPosition.Z())
			transformMatrix := entities.WorldTransform(e)
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
