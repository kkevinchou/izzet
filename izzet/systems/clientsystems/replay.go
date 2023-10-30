package clientsystems

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/systems"
	"github.com/kkevinchou/izzet/izzet/systems/shared"
)

func replay(entity *entities.Entity, gamestateUpdateMessage network.GameStateUpdateMessage, cfHistory *CommandFrameHistory, world systems.GameWorld) error {
	// cfHistory.ClearUntilFrameNumber(gamestateUpdateMessage.LastInputCommandFrame)

	// entities := []*entities.Entity{}

	// fetch all command frame inputs

	commandFrames, err := cfHistory.GetAllFramesStartingFrom(gamestateUpdateMessage.LastInputCommandFrame)
	if err != nil {
		return err
	}
	cfHistory.Reset()

	if gamestateUpdateMessage.LastInputCommandFrame != commandFrames[0].FrameNumber {
		panic("the first frame we fetch should match the last input command frame")
	}

	entities.SetLocalPosition(entity, commandFrames[0].PostCFState.Position)
	entities.SetLocalRotation(entity, commandFrames[0].PostCFState.Orientation)
	entity.Physics.Velocity = commandFrames[0].PostCFState.Velocity
	cfHistory.AddCommandFrame(gamestateUpdateMessage.LastInputCommandFrame, commandFrames[0].FrameInput, entity)

	if len(commandFrames) == 1 {
		return nil
	}

	// start loop
	//		load entity transforms of non-predicted entities
	//		set player input
	// 		simulate frame and add the new cf to the history

	for i := 1; i < len(commandFrames); i++ {
		commandFrame := commandFrames[i]

		// reset entity positions, (if they exist on the client)
		// rerun spatial partioning over these entities

		// update character controller
		// update physics

		// add a new command frame
		shared.UpdateCharacterController(time.Duration(settings.MSPerCommandFrame)*time.Millisecond, world, commandFrame.FrameInput, nil, entity)
		cfHistory.AddCommandFrame(commandFrame.FrameNumber, commandFrame.FrameInput, entity)
	}
	return nil
}
