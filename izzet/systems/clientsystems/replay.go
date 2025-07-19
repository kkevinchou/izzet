package clientsystems

import (
	"fmt"
	"time"

	"github.com/kkevinchou/izzet/izzet/apputils"
	"github.com/kkevinchou/izzet/izzet/collisionobserver"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/systems"
	"github.com/kkevinchou/izzet/izzet/systems/shared"
)

func replay(app App, entity *entities.Entity, gamestateUpdateMessage network.GameStateUpdateMessage, cfHistory *CommandFrameHistory, world systems.GameWorld) error {
	commandFrames, err := cfHistory.GetAllFramesStartingFrom(gamestateUpdateMessage.LastInputCommandFrame)
	if err != nil {
		return err
	}

	if gamestateUpdateMessage.LastInputCommandFrame != commandFrames[0].FrameNumber {
		panic("the first frame we fetch should match the last input command frame")
	}

	for _, transform := range gamestateUpdateMessage.EntityStates {
		// special case for the player for now
		if transform.EntityID != entity.GetID() {
			continue
		}

		entities.SetLocalPosition(entity, transform.Position)
		entities.SetLocalRotation(entity, transform.Rotation)
		entity.Physics.Velocity = transform.Velocity
		entity.Physics.GravityEnabled = transform.GravityEnabled

		if app.PredictionDebugLogging() {
			fmt.Printf("\t - Intialized Entity [Current Frame: %d] [Replay Frame: %d] [Position: %s]\n", app.CommandFrame(), gamestateUpdateMessage.LastInputCommandFrame, apputils.FormatVec(transform.Position))
		}
	}

	cfHistory.Reset()
	cfHistory.AddCommandFrame(gamestateUpdateMessage.LastInputCommandFrame, commandFrames[0].FrameInput, entity)

	if len(commandFrames) == 1 {
		return nil
	}

	// TODO: make this a dummy physics observer
	observer := collisionobserver.NewCollisionObserver()
	for i := 1; i < len(commandFrames); i++ {
		commandFrame := commandFrames[i]

		// reset entity positions, (if they exist on the client)
		// rerun spatial partioning over these entities ?

		shared.UpdateCharacterController(time.Duration(settings.MSPerCommandFrame)*time.Millisecond, world, commandFrame.FrameInput, entity)
		shared.PhysicsStepSingle(time.Duration(settings.MSPerCommandFrame)*time.Millisecond, entity)
		shared.ResolveCollisions(app, observer)
		if app.PredictionDebugLogging() {
			fmt.Printf("\t - Replayed Frame [Current Frame: %d] [Replay Frame: %d] [Position: %s]\n", app.CommandFrame(), commandFrame.FrameNumber, apputils.FormatVec(entity.Position()))
		}
		cfHistory.AddCommandFrame(commandFrame.FrameNumber, commandFrame.FrameInput, entity)
	}
	return nil
}
