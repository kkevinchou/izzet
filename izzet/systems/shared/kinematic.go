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
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/types"
)

func KinematicStepSingle(delta time.Duration, entity types.KinematicEntity, world GameWorld, app App) {
	KinematicStep(delta, []types.KinematicEntity{entity}, world, app)
}

func KinematicStep[T types.KinematicEntity](delta time.Duration, ents []T, world GameWorld, app App) {
	for _, e1 := range ents {
		if e1.IsStatic() || !e1.IsKinematic() {
			continue
		}

		if e1.GravityEnabled() {
			velocityFromGravity := mgl64.Vec3{0, -settings.AccelerationDueToGravity * float64(delta.Milliseconds()) / 1000}
			e1.AccumulateKinematicVelocity(velocityFromGravity)
		}

		if e1.GravityEnabled() {
			v := e1.TotalKinematicVelocity()
			vWithoutY := mgl64.Vec3{v.X(), 0, v.Z()}
			if vWithoutY != apputils.ZeroVec {
				vWithoutY = vWithoutY.Normalize()
				rotateEntityToFaceMovement(e1, vWithoutY)
			}
		}

		e1.AddPosition(e1.TotalKinematicVelocity().Mul(delta.Seconds()))

		maxRunCount := 100
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

			if minContact.SeparatingVector.Normalize().Dot(mgl64.Vec3{0, 1, 0}) > GroundedThreshold {
				grounded = true
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

func collideKinematicEntities(e1, e2 types.KinematicEntity) []collision.Contact {
	var result []collision.Contact

	if (e1.HasCapsuleCollider() && e2.HasTriMeshCollider()) || (e2.HasCapsuleCollider() && e1.HasTriMeshCollider()) {
		var capsuleCollider collider.Capsule
		var triMeshCollider collider.TriMesh

		if e1.HasCapsuleCollider() {
			capsuleCollider = e1.CapsuleCollider()
			triMeshCollider = e2.TriMeshCollider()
		} else {
			capsuleCollider = e2.CapsuleCollider()
			triMeshCollider = e1.TriMeshCollider()
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
				Type:               contact.Type,
				SeparatingVector:   contact.SeparatingVector,
				SeparatingDistance: contact.SeparatingDistance,
			}
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

		result = append(result, collision.Contact{
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
