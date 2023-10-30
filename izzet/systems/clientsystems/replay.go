package clientsystems

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/observers"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/systems"
	"github.com/kkevinchou/izzet/izzet/systems/shared"
)

func replay(entity *entities.Entity, gamestateUpdateMessage network.GameStateUpdateMessage, cfHistory *CommandFrameHistory, world systems.GameWorld) error {
	commandFrames, err := cfHistory.GetAllFramesStartingFrom(gamestateUpdateMessage.LastInputCommandFrame)
	if err != nil {
		return err
	}

	if gamestateUpdateMessage.LastInputCommandFrame != commandFrames[0].FrameNumber {
		panic("the first frame we fetch should match the last input command frame")
	}

	for _, transform := range gamestateUpdateMessage.Transforms {
		// special case for the player for now
		if transform.EntityID != entity.GetID() {
			continue
		}
		entities.SetLocalPosition(entity, transform.Position)
		entities.SetLocalRotation(entity, transform.Orientation)
		entity.Physics.Velocity = transform.Velocity
	}

	cfHistory.Reset()
	cfHistory.AddCommandFrame(gamestateUpdateMessage.LastInputCommandFrame, commandFrames[0].FrameInput, entity)

	if len(commandFrames) == 1 {
		return nil
	}

	// TODO: make this a dummy physics observer
	observer := observers.NewCollisionObserver()
	for i := 1; i < len(commandFrames); i++ {
		commandFrame := commandFrames[i]

		// reset entity positions, (if they exist on the client)
		// rerun spatial partioning over these entities ?

		shared.UpdateCharacterController(time.Duration(settings.MSPerCommandFrame)*time.Millisecond, world, commandFrame.FrameInput, entity)
		shared.PhysicsStepSingle(time.Duration(settings.MSPerCommandFrame)*time.Millisecond, entity)
		shared.ResolveCollisionsSingle(world, entity, observer)
		cfHistory.AddCommandFrame(commandFrame.FrameNumber, commandFrame.FrameInput, entity)
	}
	return nil
}
