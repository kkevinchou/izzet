package systems

import (
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/apputils"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/settings"
)

type CameraTargetSystem struct {
}

func (s *CameraTargetSystem) Name() string {
	return "CameraTargetSystem"
}

func (s *CameraTargetSystem) Update(delta time.Duration, world GameWorld) {
	for _, entity := range world.Entities() {
		if entity.CameraComponent == nil {
			continue
		}

		update(delta, world, entity)
	}
}

func update(delta time.Duration, world GameWorld, camera *entities.Entity) {
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

	entities.SetLocalPosition(camera, cameraPosition)
}
