package clientsystems

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/settings"
)

type StateBuffer struct {
	bufferedInterpolations  [settings.MaxStateBufferSize]BufferedInterpolation
	lastGameStateUpdate     network.GameStateUpdateMessage
	lastGameStateLocalFrame int
	cursor                  int
	count                   int
}

type BufferedInterpolation struct {
	CommandFrame   int
	BufferedStates []BufferedState
}

type BufferedState struct {
	EntityID int
	Position mgl64.Vec3
	Rotation mgl64.Quat
	Deadge   bool
}

func NewStateBuffer() *StateBuffer {
	return &StateBuffer{lastGameStateUpdate: network.GameStateUpdateMessage{LastInputCommandFrame: -1}}
}

func init() {
	Log = map[int]bool{}
}

var Log map[int]bool

func (sb *StateBuffer) Push(gamestateUpdateMessage network.GameStateUpdateMessage, localCommandFrame int) {
	if sb.lastGameStateUpdate.LastInputCommandFrame == -1 {
		sb.lastGameStateUpdate = gamestateUpdateMessage
		sb.lastGameStateLocalFrame = localCommandFrame
		return
	}

	blendEnd := map[int]BufferedState{}
	blendStart := map[int]BufferedState{}
	entityIDs := []int{}

	for _, entity := range gamestateUpdateMessage.EntityStates {
		// if entity.EntityID >= 5160 && !Log[entity.EntityID] {
		// 	fmt.Println("STATE BUFFER PUSH", entity.EntityID)
		// }
		blendEnd[entity.EntityID] = BufferedState{
			EntityID: entity.EntityID,
			Position: entity.Position,
			Rotation: entity.Rotation,
		}
	}

	if sb.count >= 1 {
		for _, entity := range sb.bufferedInterpolations[sb.cursor].BufferedStates {
			blendStart[entity.EntityID] = BufferedState{
				EntityID: entity.EntityID,
				Position: entity.Position,
				Rotation: entity.Rotation,
			}
		}
		sb.count = 1
	} else {
		for _, entity := range sb.lastGameStateUpdate.EntityStates {
			blendStart[entity.EntityID] = BufferedState{
				Position: entity.Position,
				Rotation: entity.Rotation,
			}
		}
	}

	// case where entity exists in the current game state update, but not the previous one
	for _, endEntity := range blendEnd {
		if _, ok := blendStart[endEntity.EntityID]; !ok {
			blendStart[endEntity.EntityID] = blendEnd[endEntity.EntityID]
		}
	}

	// case where entity exists in the last game state update, but not the current one
	for _, startEntity := range blendStart {
		if _, ok := blendEnd[startEntity.EntityID]; !ok {
			blendEnd[startEntity.EntityID] = blendStart[startEntity.EntityID]
		}
	}

	for _, entity := range blendStart {
		entityIDs = append(entityIDs, entity.EntityID)
	}

	sb.writeInterpolatedStates(gamestateUpdateMessage, blendStart, blendEnd, entityIDs)

	sb.lastGameStateUpdate = gamestateUpdateMessage
	sb.lastGameStateLocalFrame = localCommandFrame
}

func (sb *StateBuffer) writeInterpolatedStates(gamestateUpdateMessage network.GameStateUpdateMessage, blendStart map[int]BufferedState, blendEnd map[int]BufferedState, entityIDs []int) {
	numStates := gamestateUpdateMessage.GlobalCommandFrame - sb.lastGameStateUpdate.GlobalCommandFrame + 1
	cfStep := float64(1) / float64(numStates)

	for i := 1; i <= numStates; i++ {
		bi := BufferedInterpolation{CommandFrame: 0}

		for _, id := range entityIDs {
			endSnapshot := blendEnd[id]
			startSnapshot := blendStart[id]

			bs := BufferedState{
				EntityID: id,
				Position: endSnapshot.Position.Sub(startSnapshot.Position).Mul(float64(i) * cfStep).Add(startSnapshot.Position),
				Rotation: QInterpolate64(startSnapshot.Rotation, endSnapshot.Rotation, float64(i)*cfStep),
			}

			bi.BufferedStates = append(bi.BufferedStates, bs)
		}

		for _, id := range gamestateUpdateMessage.DestroyedEntities {
			bs := BufferedState{
				EntityID: id,
				Deadge:   true,
			}
			bi.BufferedStates = append(bi.BufferedStates, bs)
		}

		sb.bufferedInterpolations[(sb.cursor+sb.count)%settings.MaxStateBufferSize] = bi
		if sb.count == settings.MaxStateBufferSize {
			panic(fmt.Sprintf("buffer has filled with max capacity %d", settings.MaxStateBufferSize))
		}
		sb.count += 1
	}
}

func (sb *StateBuffer) Pull(localCommandFrame int) (BufferedInterpolation, bool) {
	if sb.count == 0 || sb.bufferedInterpolations[sb.cursor].CommandFrame > localCommandFrame {
		return BufferedInterpolation{}, false
	}

	// TODO - this should probably advance all command frames that are <= local command frame to quickly
	// catch up

	snapshot := sb.bufferedInterpolations[sb.cursor]
	sb.cursor = (sb.cursor + 1) % settings.MaxStateBufferSize
	sb.count -= 1
	return snapshot, true
}

// Quaternion interpolation, reimplemented from: https://github.com/TheThinMatrix/OpenGL-Animation/blob/dde792fe29767192bcb60d30ac3e82d6bcff1110/Animation/animation/Quaternion.java#L158
func QInterpolate64(a, b mgl64.Quat, blend float64) mgl64.Quat {
	var result mgl64.Quat = mgl64.Quat{}
	var dot float64 = a.W*b.W + a.V.X()*b.V.X() + a.V.Y()*b.V.Y() + a.V.Z()*b.V.Z()
	blendI := float64(1) - blend
	if dot < 0 {
		result.W = blendI*a.W + blend*-b.W
		result.V = mgl64.Vec3{
			blendI*a.V.X() + blend*-b.V.X(),
			blendI*a.V.Y() + blend*-b.V.Y(),
			blendI*a.V.Z() + blend*-b.V.Z(),
		}
	} else {
		result.W = blendI*a.W + blend*b.W
		result.V = mgl64.Vec3{
			blendI*a.V.X() + blend*b.V.X(),
			blendI*a.V.Y() + blend*b.V.Y(),
			blendI*a.V.Z() + blend*b.V.Z(),
		}
	}

	return result.Normalize()
}
