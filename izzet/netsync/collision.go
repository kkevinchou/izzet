package netsync

import (
	"fmt"
	"sort"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/spatialpartition"
	"github.com/kkevinchou/izzet/lib/collision"
)

type World interface {
	QueryEntity(componentFlags int) []entities.Entity
	GetPlayerEntity() entities.Entity
	GetEntityByID(id int) entities.Entity
	SpatialPartition() *spatialpartition.SpatialPartition
}

func ResolveCollisionsForPlayer(playerEntity entities.Entity, world World) {
	entityPairs := [][]entities.Entity{}
	entityList := world.SpatialPartition().QueryCollisionCandidates(playerEntity)
	for _, e2 := range entityList {
		if playerEntity.GetID() == e2.GetID() {
			continue
		}
		entityPairs = append(entityPairs, []entities.Entity{playerEntity, e2})
	}
	detectAndResolveCollisionsForEntityPairs(entityPairs, entityList, world)
}

func ResolveCollisions(world World) {
	entityPairs := [][]entities.Entity{}
	// pairExists stores the pairs of entities that we've already created,
	// don't create a pair for both (e1, e2) and (e2, e1), just one of them
	pairExists := map[int]map[int]bool{}
	entityList := world.SpatialPartition().AllCandidates()
	for _, e := range entityList {
		pairExists[e.GetID()] = map[int]bool{}
	}

	for _, e1 := range entityList {
		candidates := world.SpatialPartition().QueryCollisionCandidates(e1)
		for _, e2 := range candidates {
			if e1.GetID() == e2.GetID() {
				continue
			}

			if pairExists[e1.GetID()][e2.GetID()] || pairExists[e2.GetID()][e1.GetID()] {
				continue
			}

			entityPairs = append(entityPairs, []entities.Entity{e1, e2})
			pairExists[e1.GetID()][e2.GetID()] = true
			pairExists[e2.GetID()][e1.GetID()] = true
		}
	}
	detectAndResolveCollisionsForEntityPairs(entityPairs, entityList, world)
}

func detectAndResolveCollisionsForEntityPairs(entityPairs [][]entities.Entity, entityList []entities.Entity, world World) {
	// 1. collect pairs of entities that are colliding, sorted by separating vector
	// 2. perform collision resolution for any colliding entities
	// 3. this can cause more collisions, repeat until no more further detected collisions, or we hit the configured max

	positionalResolutionEntityPairs := [][]entities.Entity{}
	nonPositionalResolutionEntityPairs := [][]entities.Entity{}

	for _, pair := range entityPairs {
		cc1 := pair[0].GetComponentContainer()
		cc2 := pair[1].GetComponentContainer()

		if cc1.ColliderComponent.SkipSeparation || cc2.ColliderComponent.SkipSeparation {
			nonPositionalResolutionEntityPairs = append(nonPositionalResolutionEntityPairs, pair)
		} else {
			positionalResolutionEntityPairs = append(positionalResolutionEntityPairs, pair)
		}
	}

	resolveCount := map[int]int{}
	maximallyCollidingEntities := map[int]bool{}
	absoluteMaxRunCount := len(entityList) * resolveCountMax

	// collisionRuns acts as a fail safe for when we run collision detection and resolution infinitely.
	// given that we have a cap on collision resolution for each entity, we should never run more than
	// the number of entities times the cap.
	collisionRuns := 0
	for collisionRuns = 0; collisionRuns < absoluteMaxRunCount; collisionRuns++ {
		collisionCandidates := collectSortedCollisionCandidates(positionalResolutionEntityPairs, entityList, maximallyCollidingEntities, world)
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
			e1.GetComponentContainer().ColliderComponent.Contacts[e2.GetID()] = true
			e2.GetComponentContainer().ColliderComponent.Contacts[e1.GetID()] = true
		}
	}

	if collisionRuns == absoluteMaxRunCount {
		fmt.Println("hit absolute max collision run count")
	}

	// handle entities that we skip separation for. i.e. these entities just want to know if they've collided with something
	// but it don't want its positon changed
	collisionCandidates := collectSortedCollisionCandidates(nonPositionalResolutionEntityPairs, entityList, map[int]bool{}, world)
	for _, candidate := range collisionCandidates {
		e1 := world.GetEntityByID(*candidate.EntityID)
		e2 := world.GetEntityByID(*candidate.SourceEntityID)

		// TODO: consider that two of the same entity may collide twice
		// also, we may want to support colliding with individual mesh chunks of an
		// entity rather than consideration of the whole entity itself
		e1.GetComponentContainer().ColliderComponent.Contacts[e2.GetID()] = true
		e2.GetComponentContainer().ColliderComponent.Contacts[e1.GetID()] = true
	}
}

// collectSortedCollisionCandidates collects all potential collisions that can occur in the frame.
// these are "candidates" in that they are not guaranteed to have actually happened since
// if we resolve some of the collisions in the list, some will be invalidated
func collectSortedCollisionCandidates(entityPairs [][]entities.Entity, entityList []entities.Entity, skipEntitySet map[int]bool, world World) []*collision.Contact {
	// initialize collision state
	for _, e := range entityList {
		cc := e.GetComponentContainer()
		if cc.ColliderComponent.CapsuleCollider != nil {
			capsule := cc.ColliderComponent.CapsuleCollider.Transform(cc.TransformComponent.Position)
			cc.ColliderComponent.TransformedCapsuleCollider = &capsule
		} else if cc.ColliderComponent.TriMeshCollider != nil {
			transformMatrix := mgl64.Translate3D(cc.TransformComponent.Position.X(), cc.TransformComponent.Position.Y(), cc.TransformComponent.Position.Z())
			triMesh := cc.ColliderComponent.TriMeshCollider.Transform(transformMatrix)
			cc.ColliderComponent.TransformedTriMeshCollider = &triMesh
		}
	}

	var allContacts []*collision.Contact
	for _, pair := range entityPairs {
		if _, ok := skipEntitySet[pair[0].GetID()]; ok {
			continue
		}
		if _, ok := skipEntitySet[pair[1].GetID()]; ok {
			continue
		}

		contacts := collide(pair[0], pair[1])
		if len(contacts) <= 0 {
			continue
		}

		allContacts = append(allContacts, contacts...)
	}
	sort.Sort(collision.ContactsBySeparatingDistance(allContacts))

	return allContacts
}

func collide(e1 entities.Entity, e2 entities.Entity) []*collision.Contact {
	var result []*collision.Contact

	if ok, capsuleEntity, triMeshEntity := isCapsuleTriMeshCollision(e1, e2); ok {
		contacts := collision.CheckCollisionCapsuleTriMesh(
			*capsuleEntity.GetComponentContainer().ColliderComponent.TransformedCapsuleCollider,
			*triMeshEntity.GetComponentContainer().ColliderComponent.TransformedTriMeshCollider,
		)
		if len(contacts) == 0 {
			return nil
		}

		triEntityID := triMeshEntity.GetID()
		capsuleEntityID := capsuleEntity.GetID()

		for _, contact := range contacts {
			contact.EntityID = &capsuleEntityID
			contact.SourceEntityID = &triEntityID
		}

		result = contacts
	} else if ok := isCapsuleCapsuleCollision(e1, e2); ok {
		contact := collision.CheckCollisionCapsuleCapsule(
			*e1.GetComponentContainer().ColliderComponent.TransformedCapsuleCollider,
			*e2.GetComponentContainer().ColliderComponent.TransformedCapsuleCollider,
		)
		if contact == nil {
			return nil
		}

		e1ID := e1.GetID()
		e2ID := e2.GetID()
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

func resolveCollision(entity entities.Entity, sourceEntity entities.Entity, contact *collision.Contact) {
	if contact.Type == collision.ContactTypeCapsuleTriMesh {
		cc := entity.GetComponentContainer()
		transformComponent := cc.TransformComponent
		tpcComponent := cc.ThirdPersonControllerComponent
		physicsComponent := cc.PhysicsComponent
		movementComponent := cc.MovementComponent

		separatingVector := contact.SeparatingVector
		if separatingVector.Normalize().Dot(mgl64.Vec3{0, 1, 0}) >= groundedStrictness {
			if movementComponent != nil {
				movementComponent.Velocity[1] = 0
			}
		}

		if tpcComponent != nil {
			if separatingVector.Normalize().Dot(mgl64.Vec3{0, 1, 0}) >= groundedStrictness {
				// prevent sliding when grounded
				separatingVector[0] = 0
				separatingVector[2] = 0

				tpcComponent.BaseVelocity[1] = 0
				tpcComponent.ZipVelocity = mgl64.Vec3{}
				tpcComponent.Grounded = true
			}
		} else if physicsComponent != nil {
			if separatingVector.Normalize().Dot(mgl64.Vec3{0, 1, 0}) >= groundedStrictness {
				physicsComponent.Grounded = true
				physicsComponent.Velocity[1] = 0
			}
		}
		transformComponent.Position = transformComponent.Position.Add(separatingVector)
	} else if contact.Type == collision.ContactTypeCapsuleCapsule {
		bothEntities := []entities.Entity{entity, sourceEntity}
		for _, e := range bothEntities {
			cc := e.GetComponentContainer()
			transformComponent := cc.TransformComponent
			tpcComponent := cc.ThirdPersonControllerComponent
			movementComponent := cc.MovementComponent

			if tpcComponent != nil {
				separatingVector := contact.SeparatingVector
				if e.GetID() == sourceEntity.GetID() {
					separatingVector = separatingVector.Mul(-1)
				}

				transformComponent.Position = transformComponent.Position.Add(separatingVector)
				if separatingVector.Normalize().Dot(mgl64.Vec3{0, 1, 0}) >= groundedStrictness {
					tpcComponent.Grounded = true
					movementComponent.Velocity[1] = 0
					tpcComponent.BaseVelocity[1] = 0
					tpcComponent.ZipVelocity = mgl64.Vec3{}
				}
			}
		}
	}
}

func CollisionBookKeeping(entity entities.Entity) {
	cc := entity.GetComponentContainer()
	if len(cc.ColliderComponent.Contacts) == 0 {
		if cc.ThirdPersonControllerComponent != nil {
			cc.ThirdPersonControllerComponent.Grounded = false
		}
	}
	cc.ColliderComponent.Contacts = map[int]bool{}
}

func isCapsuleTriMeshCollision(e1, e2 entities.Entity) (bool, entities.Entity, entities.Entity) {
	e1cc := e1.GetComponentContainer()
	e2cc := e2.GetComponentContainer()

	if e1cc.ColliderComponent.CapsuleCollider != nil {
		if e2cc.ColliderComponent.TriMeshCollider != nil {
			return true, e1, e2
		}
	}

	if e2cc.ColliderComponent.CapsuleCollider != nil {
		if e1cc.ColliderComponent.TriMeshCollider != nil {
			return true, e2, e1
		}
	}

	return false, nil, nil
}

func isCapsuleCapsuleCollision(e1, e2 entities.Entity) bool {
	e1cc := e1.GetComponentContainer()
	e2cc := e2.GetComponentContainer()

	if e1cc.ColliderComponent.CapsuleCollider != nil {
		if e2cc.ColliderComponent.CapsuleCollider != nil {
			return true
		}
	}

	return false
}
