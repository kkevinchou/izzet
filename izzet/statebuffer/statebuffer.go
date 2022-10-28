package statebuffer

import (
	"fmt"

	"github.com/kkevinchou/izzet/izzet/events"
	"github.com/kkevinchou/izzet/izzet/knetwork"
	"github.com/kkevinchou/izzet/lib/libutils"
)

type BufferedState struct {
	InterpolatedEntities map[int]knetwork.EntitySnapshot
	Events               []knetwork.Event
}

type IncomingEntityUpdate struct {
	targetCommandFrame     int
	globalCommandFrame     int
	gameStateUpdateMessage *knetwork.GameStateUpdateMessage
}

type StateBuffer struct {
	maxStateBufferCommandFrames int
	timeline                    map[int]BufferedState
	incomingEntityUpdates       []IncomingEntityUpdate
}

func NewStateBuffer(maxStateBufferCommandFrames int) *StateBuffer {
	return &StateBuffer{
		maxStateBufferCommandFrames: maxStateBufferCommandFrames,
		timeline:                    map[int]BufferedState{},
	}
}

func (s *StateBuffer) PushEntityUpdate(localCommandFrame int, gameStateUpdateMessage *knetwork.GameStateUpdateMessage) {
	targetCF := localCommandFrame + s.maxStateBufferCommandFrames + 1

	if len(s.incomingEntityUpdates) == 0 {
		s.incomingEntityUpdates = append(
			s.incomingEntityUpdates,
			IncomingEntityUpdate{
				gameStateUpdateMessage: gameStateUpdateMessage,
				targetCommandFrame:     targetCF,
				globalCommandFrame:     gameStateUpdateMessage.CurrentGlobalCommandFrame,
			},
		)

		s.timeline[targetCF] = BufferedState{
			InterpolatedEntities: gameStateUpdateMessage.Entities,
		}

		return
	}

	lastEntityUpdate := s.incomingEntityUpdates[len(s.incomingEntityUpdates)-1]
	currentEntityUpdate := IncomingEntityUpdate{
		gameStateUpdateMessage: gameStateUpdateMessage,
		targetCommandFrame:     targetCF,
		globalCommandFrame:     gameStateUpdateMessage.CurrentGlobalCommandFrame,
	}

	s.incomingEntityUpdates = append(
		s.incomingEntityUpdates,
		currentEntityUpdate,
	)

	s.generateIntermediateStateUpdates(lastEntityUpdate, currentEntityUpdate)
	s.incomingEntityUpdates = s.incomingEntityUpdates[1:]
}

// TODO: move interpolation logic in stateinterpolator system?
func (s *StateBuffer) generateIntermediateStateUpdates(start IncomingEntityUpdate, end IncomingEntityUpdate) {
	delta := end.targetCommandFrame - start.targetCommandFrame
	cfStep := float64(1) / float64(delta)

	// TODO(kevin) doing deserialization here is probably omegaslow

	unregisteredEntities := map[int]any{}
	for _, event := range end.gameStateUpdateMessage.Events {
		if event.Type == events.EventTypeUnregisterEntity {
			var e events.UnregisterEntityEvent
			knetwork.Deserialize(event.Bytes, &e)
			unregisteredEntities[e.EntityID] = true
		}
	}

	// if _, ok := start.gameStateUpdateMessage.Entities[80007]; ok {
	// 	fmt.Println(start.gameStateUpdateMessage.Entities[80007].Components)
	// }

	for i := 1; i <= delta; i++ {
		interpolatedEntities := map[int]knetwork.EntitySnapshot{}

		for id, startSnapshot := range start.gameStateUpdateMessage.Entities {
			if _, ok := end.gameStateUpdateMessage.Entities[id]; !ok {
				// an entity may be deleted in between two game state updates.
				if _, ok := unregisteredEntities[id]; !ok {
					fmt.Printf("warning, entity from start update (%d) did not exist in the next one and a deletion event for it was not found\n", id)
				}

				// drop the entity at the last cf
				// TODO(kevin) use the gcf to determine which frame we should drop the entity
				if i != delta {
					interpolatedEntities[id] = knetwork.EntitySnapshot{
						ID:          startSnapshot.ID,
						Type:        startSnapshot.Type,
						Position:    startSnapshot.Position,
						Orientation: startSnapshot.Orientation,
						Velocity:    startSnapshot.Velocity,
						Animation:   startSnapshot.Animation,
						Components:  startSnapshot.Components,
					}
				}
			} else {
				endSnapshot := end.gameStateUpdateMessage.Entities[id]
				referenceSnapshot := startSnapshot
				if i == delta {
					referenceSnapshot = endSnapshot
				}
				interpolatedEntities[id] = knetwork.EntitySnapshot{
					ID:          referenceSnapshot.ID,
					Type:        referenceSnapshot.Type,
					Position:    endSnapshot.Position.Sub(startSnapshot.Position).Mul(float64(i) * cfStep).Add(startSnapshot.Position),
					Orientation: libutils.QInterpolate64(startSnapshot.Orientation, endSnapshot.Orientation, float64(i)*cfStep),
					Velocity:    endSnapshot.Velocity.Sub(startSnapshot.Velocity).Mul(float64(i) * cfStep).Add(startSnapshot.Velocity),
					Animation:   referenceSnapshot.Animation,
					Components:  referenceSnapshot.Components,
				}
			}
		}

		bufferedState := BufferedState{
			InterpolatedEntities: interpolatedEntities,
		}

		if i == delta {
			bufferedState.Events = end.gameStateUpdateMessage.Events
		}

		s.timeline[start.targetCommandFrame+i] = bufferedState
	}
	// lastState := s.timeline[start.targetCommandFrame+delta]
	// lastState.Events = end.gameStateUpdateMessage.Events
}

func (s *StateBuffer) PeekEntityInterpolations(cf int) *BufferedState {
	if b, ok := s.timeline[cf]; ok {
		return &b
	}
	return nil
}

func (s *StateBuffer) PullEntityInterpolations(cf int) *BufferedState {
	if b := s.PeekEntityInterpolations(cf); b != nil {
		delete(s.timeline, cf)
		return b
	}
	return nil
}
