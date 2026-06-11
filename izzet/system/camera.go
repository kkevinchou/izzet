package system

import (
	"math"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/collision/checks"
	"github.com/kkevinchou/izzet/internal/collision/collider"
	"github.com/kkevinchou/izzet/izzet/apputils"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/settings"
)

type CameraSystem struct {
	app App
}

func NewCameraSystem(app App) *CameraSystem {
	return &CameraSystem{app: app}
}

func (s *CameraSystem) Name() string {
	return "CameraSystem"
}

func (s *CameraSystem) Update(delta time.Duration, world GameWorld) {
	for _, entity := range world.Entities() {
		if entity.CameraComponent == nil {
			continue
		}

		s.setFOVX(delta, entity)
		s.update(delta, world, entity)
	}
}

func (s *CameraSystem) setFOVX(delta time.Duration, camera *entity.Entity) {
	if s.app.IsClient() {
		camera.CameraComponent.FovX = float64(s.app.RuntimeConfig().FovX)
	} else {
		camera.CameraComponent.FovX = settings.DefaultFOVX
	}
}

func (s *CameraSystem) update(delta time.Duration, world GameWorld, camera *entity.Entity) {
	// swivel around target
	target := world.GetEntityByID(camera.CameraComponent.Target)
	if target == nil {
		return
	}

	position := target.Position()
	if target.RenderBlend.Active {
		deltaMs := time.Since(target.RenderBlend.StartTime).Milliseconds()
		t := apputils.RenderBlendMath(deltaMs)
		position = position.Sub(target.RenderBlend.BlendStartPosition).Mul(t).Add(target.RenderBlend.BlendStartPosition)
	}

	// pivot is the 3d point that the camera will rotate around
	pivot := position.Add(mgl64.Vec3{0, 1.75, 0})

	// vecFromPivot is a relative vector from the pivot to the camera
	// this vector is not yet in world space
	var vecFromPivot mgl64.Vec3
	if camera.CameraComponent.CameraMode == entity.CameraModeWideView {
		vecFromPivot = mgl64.Vec3{0, 0, 5}
	} else if camera.CameraComponent.CameraMode == entity.CameraModeOverShoulder {
		if target.AimDownSightsComponent != nil && target.AimDownSightsComponent.Active {
			vecFromPivot = mgl64.Vec3{0.6, 0, 1.1}
			camera.CameraComponent.FovX = 85
		} else {
			vecFromPivot = mgl64.Vec3{0.65, 0, 2.0}
		}
	} else {
		panic("wat")
	}

	cameraWorldSpacePosition := camera.GetLocalRotation().Rotate(vecFromPivot).Add(pivot)
	entityCameraLine := collider.Line{P1: pivot, P2: cameraWorldSpacePosition}
	ents := s.app.World().SpatialPartition().EntitiesByLineSegment(entityCameraLine)

	dir := cameraWorldSpacePosition.Sub(pivot)
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
		if ent.ID == target.GetID() {
			continue
		}

		ray := collider.Ray{Origin: pivot, Direction: dir}

		if _, _, success := checks.IntersectLineAABB(entityCameraLine, ent.BoundingBox()); !success {
			continue
		}

		point, _, success := checks.IntersectRayTriMesh(ray, ent.TriMeshCollider())
		if !success {
			continue
		}

		distSq := point.Sub(pivot).LenSqr()
		if distSq < minDistSq && distSq < cameraDistanceSqr {
			hit = true
			minDistSq = distSq
			hitPoint = point
		}
	}

	if hit {
		entity.SetLocalPosition(camera, hitPoint)
	} else {
		entity.SetLocalPosition(camera, cameraWorldSpacePosition)
	}
}
