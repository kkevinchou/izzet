package shared

import (
	"fmt"
	"math"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/collision"
	"github.com/kkevinchou/izzet/internal/collision/checks"
	"github.com/kkevinchou/izzet/internal/collision/collider"
	"github.com/kkevinchou/izzet/internal/utils"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/types"
)

const maxRunCount int = 100
const maxSlopeAngle float64 = 45
const maxSlopeRadians float64 = maxSlopeAngle * math.Pi / 180
const groundedStickDistance float64 = 0.05

type RayCastResult struct {
	normal      mgl64.Vec3
	hitDistance float64
	hit         bool
}

func KinematicStepSingle(delta time.Duration, e *entity.Entity, world GameWorld, app App) {
	KinematicStep(delta, []*entity.Entity{e}, world, app)
}

func KinematicStep(delta time.Duration, ents []*entity.Entity, world GameWorld, app App) {
	for _, e1 := range ents {
		if e1.IsStatic() || !e1.IsKinematic() {
			continue
		}

		if e1.Kinematic.Jump {
			e1.Kinematic.Grounded = false
			jumpVelocity := mgl64.Vec3{0, settings.CharacterJumpSpeed, 0}
			e1.Kinematic.AccumulatedVelocity = e1.Kinematic.AccumulatedVelocity.Add(jumpVelocity)
		}

		movementDir := e1.Kinematic.MoveIntent

		if e1.GravityEnabled() {
			// disable Y movements when gravity is enabled
			movementDir = removeYMovement(movementDir)
			groundRayCastResult := rayCastToGround(world, e1)
			if groundRayCastResult.hit {
				movementDir = moveDirAlongSlope(movementDir, groundRayCastResult.normal)
			}

			_, _, supported := walkableGroundSupportFromProbe(e1, groundRayCastResult)

			// only apply gravity if we aren't supported by the ground or if we have a vertical velocity component
			if !supported || e1.Kinematic.AccumulatedVelocity.Y() > 0 {
				velocityFromGravity := mgl64.Vec3{0, -settings.AccelerationDueToGravity * float64(delta.Milliseconds()) / 1000}
				e1.AccumulateKinematicVelocity(velocityFromGravity)
			} else {
				// Keep the character pinned in place on walkable slopes while idle.
				// Preserve upward velocity so jump impulses are not canceled.
				if e1.Kinematic.AccumulatedVelocity.Y() < 0 {
					e1.ClearVerticalKinematicVelocity()
				}
			}
		}

		e1.Kinematic.Velocity = movementDir.Mul(e1.Kinematic.Speed)
		e1.AddPosition(e1.TotalKinematicVelocity().Mul(delta.Seconds()))

		v := e1.TotalKinematicVelocity()
		vWithoutY := mgl64.Vec3{v.X(), 0, v.Z()}
		if !utils.Vec3IsZero(vWithoutY) {
			vWithoutY = vWithoutY.Normalize()
			rotateEntityToFaceMovement(e1, vWithoutY)
		}

		var runCount int = 0

		var grounded bool
		for runCount = range maxRunCount {
			e1BoundingBox := e1.BoundingBox()
			candidates := world.SpatialPartition().QueryEntities(e1BoundingBox)

			if len(candidates) == 0 {
				break
			}

			var minDist float64 = math.MaxFloat64
			var minContact collision.Contact

			for _, partitionEntity := range candidates {
				var e2 types.KinematicEntity = world.GetEntityByID(partitionEntity.GetID())
				if e1.GetID() == e2.GetID() {
					continue
				}

				if !checks.BoundingBoxOverlaps(e1BoundingBox, e2.BoundingBox()) {
					continue
				}

				contacts := collideKinematicEntities(e1, e2)
				for _, contact := range contacts {
					if contact.SeparatingDistance < minDist {
						minDist = contact.SeparatingDistance
						minContact = contact
					}
				}
			}

			if minDist == math.MaxFloat64 {
				break
			}

			if e1.GravityEnabled() {
				slopeRadians := slopeAngleFromNormal(minContact.SeparatingVector)
				if slopeRadians <= maxSlopeRadians {
					grounded = true
				}
			}

			e1.AddPosition(minContact.SeparatingVector)
		}

		// Maintain grounding when standing still in resting contact (no overlap).
		if e1.GravityEnabled() && !grounded {
			_, _, supported := walkableGroundSupport(world, e1)
			grounded = supported
		}

		e1.SetGrounded(grounded)
		if grounded {
			e1.ClearVerticalKinematicVelocity()
		}

		if runCount == maxRunCount-1 {
			fmt.Printf("HIT KINEMATIC MAX RUNCOUNT OF %d\n", maxRunCount)
		}
	}
}

func rayCastToGround(world GameWorld, e1 types.KinematicEntity) RayCastResult {
	capsule := e1.CapsuleCollider()
	rayOrigin := capsule.Bottom
	ray := collider.Ray{Origin: rayOrigin, Direction: mgl64.Vec3{0, -1, 0}}

	result := RayCastResult{hitDistance: math.MaxFloat32}

	maxWalkableGroundDistance := capsule.Radius/math.Cos(maxSlopeRadians) + groundedStickDistance
	queryBounds := e1.BoundingBox()
	queryBounds.MinVertex = queryBounds.MinVertex.Sub(mgl64.Vec3{0, maxWalkableGroundDistance, 0})
	candidates := world.SpatialPartition().QueryEntities(queryBounds)

	for _, candidate := range candidates {
		if candidate.GetID() == e1.GetID() {
			continue
		}

		e2 := world.GetEntityByID(candidate.GetID())
		if !e2.HasTriMeshCollider() {
			continue
		}

		if !checks.BoundingBoxOverlaps(queryBounds, e2.BoundingBox()) {
			continue
		}

		var triMesh collider.TriMesh
		if e2.HasSimplifiedTriMeshCollider() {
			triMesh = e2.SimplifiedTriMeshCollider()
		} else {
			triMesh = e2.TriMeshCollider()
		}

		hitPoint, hitNormal, hit := checks.IntersectRayTriMesh(ray, triMesh)
		if !hit {
			continue
		}

		hitDistance := hitPoint.Sub(rayOrigin).Len()
		if hitDistance < result.hitDistance {
			result.normal = hitNormal
			result.hitDistance = hitDistance
			result.hit = true
		}
	}

	return result
}

func walkableGroundSupport(world GameWorld, e1 types.KinematicEntity) (mgl64.Vec3, float64, bool) {
	return walkableGroundSupportFromProbe(e1, rayCastToGround(world, e1))
}

func walkableGroundSupportFromProbe(e1 types.KinematicEntity, groundRayCastResult RayCastResult) (mgl64.Vec3, float64, bool) {
	if !groundRayCastResult.hit {
		return mgl64.Vec3{}, 0, false
	}

	normal := groundRayCastResult.normal
	hitDistance := groundRayCastResult.hitDistance
	if normal.LenSqr() == 0 {
		return mgl64.Vec3{}, 0, false
	}

	normalizedNormal := normal.Normalize()
	if slopeAngleFromNormal(normalizedNormal) > maxSlopeRadians {
		return normal, hitDistance, false
	}

	if normalizedNormal.Y() <= 0 {
		return normal, hitDistance, false
	}

	// When a sphere rests on a slope, the vertical center-to-ground distance grows as 1/normal.y.
	r := e1.CapsuleCollider().Radius
	restingDistance := r / normalizedNormal.Y()
	if hitDistance > restingDistance+groundedStickDistance {
		return normal, hitDistance, false
	}

	return normal, hitDistance, true
}

func slopeAngleFromNormal(normal mgl64.Vec3) float64 {
	if normal.LenSqr() == 0 {
		return math.Pi
	}

	dot := normal.Normalize().Dot(mgl64.Vec3{0, 1, 0})
	dot = mgl64.Clamp(dot, -1, 1)
	return math.Acos(dot)
}

// moveDirAlongSlope computes a vector representing the movement direction along a slope
func moveDirAlongSlope(movementDir mgl64.Vec3, slopeNormal mgl64.Vec3) mgl64.Vec3 {
	slopeRadians := slopeAngleFromNormal(slopeNormal)
	if slopeRadians > maxSlopeRadians {
		return movementDir
	}

	if movementDir.LenSqr() == 0 {
		return movementDir
	}

	movementDir = movementDir.Normalize()
	y := -(movementDir.X()*slopeNormal.X() + movementDir.Z()*slopeNormal.Z()) / slopeNormal.Y()
	movementDir[1] = y

	return movementDir.Normalize()
}

func collideKinematicEntities(e1, e2 types.KinematicEntity) []collision.Contact {
	var result []collision.Contact

	if (e1.HasCapsuleCollider() && e2.HasTriMeshCollider()) || (e2.HasCapsuleCollider() && e1.HasTriMeshCollider()) {
		var capsuleCollider collider.Capsule
		var triMeshCollider collider.TriMesh

		if e1.HasCapsuleCollider() {
			capsuleCollider = e1.CapsuleCollider()
			triMeshCollider = e2.TriMeshCollider()
			if e2.HasSimplifiedTriMeshCollider() {
				triMeshCollider = e2.SimplifiedTriMeshCollider()
			}
		} else {
			capsuleCollider = e2.CapsuleCollider()
			triMeshCollider = e1.TriMeshCollider()
			if e1.HasSimplifiedTriMeshCollider() {
				triMeshCollider = e1.SimplifiedTriMeshCollider()
			}
		}

		contacts := collision.CheckCollisionCapsuleTriMesh(
			capsuleCollider,
			triMeshCollider,
		)

		if len(contacts) == 0 {
			return nil
		}

		for _, contact := range contacts {
			c := contact
			if e2.HasCapsuleCollider() {
				c.SeparatingVector = c.SeparatingVector.Mul(-1)
			}
			result = append(result, c)
		}
	} else if e1.HasCapsuleCollider() && e2.HasCapsuleCollider() {
		contact, collisionDetected := collision.CheckCollisionCapsuleCapsule(
			e1.CapsuleCollider(),
			e2.CapsuleCollider(),
		)

		if !collisionDetected {
			return nil
		}

		result = append(result, contact)
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

func rotateEntityToFaceMovement(entity types.KinematicEntity, movementDirWithoutY mgl64.Vec3) {
	if !utils.Vec3IsZero(movementDirWithoutY) {
		currentRotation := entity.GetLocalRotation()
		currentViewingVector := currentRotation.Rotate(mgl64.Vec3{0, 0, -1})
		newViewingVector := movementDirWithoutY

		dot := currentViewingVector.Dot(newViewingVector)
		dot = mgl64.Clamp(dot, -1, 1)
		acuteAngle := math.Acos(dot)

		turnAnglePerFrame := (2 * math.Pi / 1000) * 2 * float64(settings.MSPerCommandFrame)

		if left := currentViewingVector.Cross(newViewingVector).Y() > 0; !left {
			turnAnglePerFrame = -turnAnglePerFrame
		}

		var newRotation mgl64.Quat

		// turning angle is less than the goal
		if math.Abs(turnAnglePerFrame) < acuteAngle {
			turningQuaternion := mgl64.QuatRotate(turnAnglePerFrame, mgl64.Vec3{0, 1, 0})
			newRotation = turningQuaternion.Mul(currentRotation)
		} else {
			// turning angle overshoots the goal, snap
			newRotation = mgl64.QuatBetweenVectors(mgl64.Vec3{0, 0, -1}, movementDirWithoutY)
		}

		entity.SetLocalRotation(newRotation)
	}
}
