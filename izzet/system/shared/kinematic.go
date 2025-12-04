package shared

import (
	"fmt"
	"math"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/collision"
	"github.com/kkevinchou/izzet/internal/collision/checks"
	"github.com/kkevinchou/izzet/internal/collision/collider"
	"github.com/kkevinchou/izzet/izzet/apputils"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/types"
)

const maxRunCount int = 100
const maxSlopeAngle float64 = 45
const maxSlopeRadians float64 = maxSlopeAngle * math.Pi / 180

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
			velocityFromGravity := mgl64.Vec3{0, -settings.AccelerationDueToGravity * float64(delta.Milliseconds()) / 1000}
			e1.AccumulateKinematicVelocity(velocityFromGravity)

			// disable Y movements when gravity is enabled
			movementDir = removeYMovement(movementDir)
			movementDir = moveDirAlongSlope(world, e1, movementDir)
		}

		e1.Kinematic.Velocity = movementDir.Mul(e1.Kinematic.Speed)
		e1.AddPosition(e1.TotalKinematicVelocity().Mul(delta.Seconds()))

		v := e1.TotalKinematicVelocity()
		vWithoutY := mgl64.Vec3{v.X(), 0, v.Z()}
		if vWithoutY != apputils.ZeroVec {
			vWithoutY = vWithoutY.Normalize()
			rotateEntityToFaceMovement(e1, vWithoutY)
		}

		var runCount int = 0

		var grounded bool
		for runCount = range maxRunCount {
			candidates := world.SpatialPartition().QueryEntities(e1.BoundingBox())

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

				if !checks.BoundingBoxOverlaps(e1.BoundingBox(), e2.BoundingBox()) {
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
				slopeRadians := math.Acos(minContact.SeparatingVector.Normalize().Dot(mgl64.Vec3{0, 1, 0}))
				if slopeRadians <= maxSlopeRadians {
					grounded = true
				}
			}

			e1.AddPosition(minContact.SeparatingVector)
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

func rayCastToGround(world GameWorld, e1 types.KinematicEntity) (mgl64.Vec3, float64, bool) {
	capsule := e1.CapsuleCollider()
	rayOrigin := capsule.Bottom.Add(mgl64.Vec3{0, -capsule.Radius, 0})
	ray := collider.Ray{Origin: rayOrigin, Direction: mgl64.Vec3{0, -1, 0}}

	var actualHit bool
	var actualHitNormal mgl64.Vec3
	var minHitDistance = math.MaxFloat32

	candidates := world.SpatialPartition().QueryEntities(e1.BoundingBox())
	for _, candidate := range candidates {
		if candidate.GetID() == e1.GetID() {
			continue
		}

		e2 := world.GetEntityByID(candidate.GetID())
		if !e2.HasTriMeshCollider() {
			continue
		}

		hitPoint, hitNormal, hit := checks.IntersectRayTriMesh(ray, e2.TriMeshCollider())
		hitDistance := hitPoint.Sub(rayOrigin).Len()
		if hit && hitDistance < minHitDistance {
			actualHitNormal = hitNormal
			minHitDistance = hitDistance
			actualHit = true
		}
	}

	return actualHitNormal, minHitDistance, actualHit
}

// moveDirAlongSlope augments the movement vector to point in the direction of any
// slopes it may be walking along
func moveDirAlongSlope(world GameWorld, e1 *entity.Entity, movementDir mgl64.Vec3) mgl64.Vec3 {
	normal, _, hit := rayCastToGround(world, e1)

	if !hit {
		return movementDir
	}

	slopeRadians := math.Acos(normal.Dot(mgl64.Vec3{0, 1, 0}))
	if slopeRadians > maxSlopeRadians {
		return movementDir
	}

	if movementDir.LenSqr() == 0 {
		return movementDir
	}

	movementDir = movementDir.Normalize()
	y := -(movementDir.X()*normal.X() + movementDir.Z()*normal.Z()) / normal.Y()
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
	if movementDirWithoutY != apputils.ZeroVec {
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
