package systems

import (
	"math"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/collision/checks"
	"github.com/kkevinchou/izzet/internal/collision/collider"
	"github.com/kkevinchou/izzet/izzet/apputils"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/settings"
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

func (s *CameraTargetSystem) update(delta time.Duration, world GameWorld, camera *entities.Entity) {
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
	targetPosition := position.Add(camera.CameraComponent.TargetPositionOffset)

	cameraOffset := settings.CameraEntityFollowDistance
	if settings.FirstPersonCamera {
		cameraOffset = 0
	}

	cameraPosition := camera.GetLocalRotation().Rotate(mgl64.Vec3{0, 0, cameraOffset}).Add(targetPosition)

	dir := cameraPosition.Sub(targetPosition)
	cameraDistanceSqr := dir.LenSqr()
	dir = dir.Normalize()

	var hit bool
	var hitPoint mgl64.Vec3

	minDistSq := math.MaxFloat64

	for _, entity := range world.Entities() {
		if entity.Collider == nil || entity.Collider.TriMeshCollider == nil {
			continue
		}
		if entity.ID == *targetID {
			continue
		}

		ray := collider.Ray{Origin: targetPosition, Direction: dir}
		transformMatrix := entities.WorldTransform(entity)
		collider := entity.Collider.TriMeshCollider.Transform(transformMatrix)

		point, success := checks.IntersectRayTriMesh(ray, collider)
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
		entities.SetLocalPosition(camera, hitPoint)
	} else {
		entities.SetLocalPosition(camera, cameraPosition)
	}
}
