package clientsystems

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/systems"
	"github.com/kkevinchou/izzet/izzet/systems/shared"
)

func replay(app App, e *entity.Entity, gamestateUpdateMessage network.GameStateUpdateMessage, cfHistory *CommandFrameHistory, world systems.GameWorld) error {
	commandFrames, err := cfHistory.GetAllFramesStartingFrom(gamestateUpdateMessage.LastInputCommandFrame)
	if err != nil {
		return err
	}

	if gamestateUpdateMessage.LastInputCommandFrame != commandFrames[0].FrameNumber {
		panic("the first frame we fetch should match the last input command frame")
	}

	for _, transform := range gamestateUpdateMessage.EntityStates {
		// special case for the player for now
		if transform.EntityID != e.GetID() {
			continue
		}

		entity.SetLocalPosition(e, transform.Position)
		e.SetLocalRotation(transform.Rotation)
		e.Kinematic.Velocity = transform.Velocity
		e.Kinematic.GravityEnabled = transform.GravityEnabled

		// if app.PredictionDebugLogging() {
		// 	fmt.Printf("\t - Intialized Entity [Current Frame: %d] [Replay Frame: %d] [Position: %s]\n", app.CommandFrame(), gamestateUpdateMessage.LastInputCommandFrame, apputils.FormatVec(transform.Position))
		// }
	}

	cfHistory.Reset()
	cfHistory.AddCommandFrame(gamestateUpdateMessage.LastInputCommandFrame, commandFrames[0].FrameInput, e)

	if len(commandFrames) == 1 {
		return nil
	}

	// TODO: make this a dummy physics observer
	// observer := collisionobserver.NewCollisionObserver()
	for i := 1; i < len(commandFrames); i++ {
		commandFrame := commandFrames[i]

		// reset entity positions, (if they exist on the client)
		// rerun spatial partioning over these entities ?

		shared.UpdateCharacterController(time.Duration(settings.MSPerCommandFrame)*time.Millisecond, commandFrame.FrameInput, e)
		shared.KinematicStepSingle(time.Duration(settings.MSPerCommandFrame)*time.Millisecond, e, app.World(), app)
		// shared.PhysicsStepSingle(time.Duration(settings.MSPerCommandFrame)*time.Millisecond, e)
		// shared.ResolveCollisions(app, observer)
		// if app.PredictionDebugLogging() {
		// 	fmt.Printf("\t - Replayed Frame [Current Frame: %d] [Replay Frame: %d] [Position: %s]\n", app.CommandFrame(), commandFrame.FrameNumber, apputils.FormatVec(e.Position()))
		// }
		cfHistory.AddCommandFrame(commandFrame.FrameNumber, commandFrame.FrameInput, e)
	}
	return nil
}
