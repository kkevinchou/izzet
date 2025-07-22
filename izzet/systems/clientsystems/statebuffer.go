package clientsystems

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/settings"
)

type StateBuffer struct {
	frames       [settings.MaxStateBufferSize]Frame
	prevGSUpdate network.GameStateUpdateMessage
	cursor       int
	count        int
}

type Frame struct {
	EntityStates []EntityState
}

type EntityState struct {
	EntityID int
	Position mgl64.Vec3
	Rotation mgl64.Quat
	Deadge   bool
}

func NewStateBuffer() *StateBuffer {
	return &StateBuffer{prevGSUpdate: network.GameStateUpdateMessage{LastInputCommandFrame: -1}}
}

func (sb *StateBuffer) Push(updateMsg network.GameStateUpdateMessage, localCommandFrame int) {
	if sb.prevGSUpdate.LastInputCommandFrame == -1 {
		sb.prevGSUpdate = updateMsg
		return
	}

	blendEnd := map[int]EntityState{}
	blendStart := map[int]EntityState{}

	for _, entity := range updateMsg.EntityStates {
		blendEnd[entity.EntityID] = EntityState{
			EntityID: entity.EntityID,
			Position: entity.Position,
			Rotation: entity.Rotation,
		}
	}

	if sb.count >= 1 {
		// if we have interpolated frames remaining, we essentially
		// toss out the remaining interpolated frames and start generating
		// new ones from the new incoming game state update

		for _, entity := range sb.frames[sb.cursor].EntityStates {
			blendStart[entity.EntityID] = EntityState{
				EntityID: entity.EntityID,
				Position: entity.Position,
				Rotation: entity.Rotation,
			}
		}

		sb.count = 1
	} else {
		for _, entity := range sb.prevGSUpdate.EntityStates {
			blendStart[entity.EntityID] = EntityState{
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

	sb.writeInterpolatedStates(updateMsg, blendStart, blendEnd)

	sb.prevGSUpdate = updateMsg
}

func (sb *StateBuffer) writeInterpolatedStates(updateMsg network.GameStateUpdateMessage, blendStart map[int]EntityState, blendEnd map[int]EntityState) {
	numFrames := updateMsg.GlobalCommandFrame - sb.prevGSUpdate.GlobalCommandFrame + 1
	cfStep := float64(1) / float64(numFrames)

	for i := 1; i <= numFrames; i++ {
		frame := Frame{}

		for id := range blendStart {
			endSnapshot := blendEnd[id]
			startSnapshot := blendStart[id]

			bs := EntityState{
				EntityID: id,
				Position: endSnapshot.Position.Sub(startSnapshot.Position).Mul(float64(i) * cfStep).Add(startSnapshot.Position),
				Rotation: QInterpolate64(startSnapshot.Rotation, endSnapshot.Rotation, float64(i)*cfStep),
			}

			frame.EntityStates = append(frame.EntityStates, bs)
		}

		for _, id := range updateMsg.DestroyedEntities {
			bs := EntityState{
				EntityID: id,
				Deadge:   true,
			}
			frame.EntityStates = append(frame.EntityStates, bs)
		}

		sb.frames[(sb.cursor+sb.count)%settings.MaxStateBufferSize] = frame
		if sb.count == settings.MaxStateBufferSize {
			panic(fmt.Sprintf("buffer has filled with max capacity %d", settings.MaxStateBufferSize))
		}
		sb.count += 1
	}
}

func (sb *StateBuffer) Pull(localCommandFrame int) (Frame, bool) {
	if sb.count == 0 {
		return Frame{}, false
	}

	frame := sb.frames[sb.cursor]
	sb.cursor = (sb.cursor + 1) % settings.MaxStateBufferSize
	sb.count -= 1
	return frame, true
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
