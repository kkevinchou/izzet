package system

import (
	"math"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/collision/checks"
	"github.com/kkevinchou/izzet/internal/collision/collider"
	"github.com/kkevinchou/izzet/izzet/apputils"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/runtimeconfig"
)

type CameraTargetSystem struct {
	app App
}

func NewCameraTargetSystem(app App) *CameraTargetSystem {
	return &CameraTargetSystem{app: app}
}

func (s *CameraTargetSystem) Name() string {
	return "CameraTargetSystem"
}

func (s *CameraTargetSystem) Update(delta time.Duration, world GameWorld) {
	for _, entity := range world.Entities() {
		if entity.CameraComponent == nil {
			continue
		}

		s.update(delta, world, entity)
	}
}

func (s *CameraTargetSystem) update(delta time.Duration, world GameWorld, camera *entity.Entity) {
	if camera.CameraComponent.Target == nil {
		return
	}

	targetID := camera.CameraComponent.Target
	if targetID == nil {
		return
	}

	targetEntity := world.GetEntityByID(*targetID)
	if targetEntity == nil {
		return
	}

	// swivel around target
	target := world.GetEntityByID(*camera.CameraComponent.Target)
	position := target.Position()
	if target.RenderBlend.Active {
		deltaMs := time.Since(target.RenderBlend.StartTime).Milliseconds()
		t := apputils.RenderBlendMath(deltaMs)
		position = position.Sub(target.RenderBlend.BlendStartPosition).Mul(t).Add(target.RenderBlend.BlendStartPosition)
	}

	var targetPosition mgl64.Vec3
	var cameraPosition mgl64.Vec3

	if s.app.IsClient() {
		s.app.RuntimeConfig().FovX = runtimeconfig.DefaultFovX
		runtimeConfig := s.app.RuntimeConfig()
		targetPosition = position.Add(runtimeConfig.CameraTargetOffset)

		cameraOffset := runtimeConfig.CameraOverShoulderOffset
		if target.AimDownSightsComponent != nil && target.AimDownSightsComponent.Active {
			cameraOffset = mgl64.Vec3{0.6, 0, 1.1}
			s.app.RuntimeConfig().FovX = 85
		}

		cameraPosition = camera.GetLocalRotation().Rotate(cameraOffset).Add(targetPosition)
		if camera.CameraComponent.CameraMode == entity.CameraModeWideView {
			cameraPosition = camera.GetLocalRotation().Rotate(mgl64.Vec3{0, 0, 5}).Add(targetPosition)
		}
	} else {
		targetPosition = position.Add(mgl64.Vec3{0, 1.75, 0})

		cameraPosition = camera.GetLocalRotation().Rotate(mgl64.Vec3{0.65, 0, 2.0}).Add(targetPosition)
		if camera.CameraComponent.CameraMode == entity.CameraModeWideView {
			cameraPosition = camera.GetLocalRotation().Rotate(mgl64.Vec3{0, 0, 5}).Add(targetPosition)
		}
	}

	entityCameraLine := collider.Line{P1: targetPosition, P2: cameraPosition}
	ents := s.app.World().SpatialPartition().EntitiesByLineSegment(entityCameraLine)

	dir := cameraPosition.Sub(targetPosition)
	cameraDistanceSqr := dir.LenSqr()
	dir = dir.Normalize()

	var hit bool
	var hitPoint mgl64.Vec3

	minDistSq := math.MaxFloat64

	for _, e := range ents {
		ent := s.app.World().GetEntityByID(e.GetID())
		if ent.Collider == nil || ent.Collider.TriMeshCollider == nil {
			continue
		}
		if ent.ID == *targetID {
			continue
		}

		ray := collider.Ray{Origin: targetPosition, Direction: dir}

		if _, _, success := checks.IntersectLineAABB(entityCameraLine, ent.BoundingBox()); !success {
			continue
		}

		point, _, success := checks.IntersectRayTriMesh(ray, ent.TriMeshCollider())
		if !success {
			continue
		}

		distSq := point.Sub(targetPosition).LenSqr()
		if distSq < minDistSq && distSq < cameraDistanceSqr {
			hit = true
			minDistSq = distSq
			hitPoint = point
		}
	}

	if hit {
		entity.SetLocalPosition(camera, hitPoint)
	} else {
		entity.SetLocalPosition(camera, cameraPosition)
	}
}
