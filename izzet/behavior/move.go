package behavior

import (
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/directory"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/izzet/lib/behavior"
	"github.com/kkevinchou/izzet/lib/geometry"
	"github.com/kkevinchou/izzet/lib/logger"
)

type Mover interface {
	types.Positionable
	SetTarget(target mgl64.Vec3)
}

type Move struct {
	Entity    Mover
	path      []geometry.Point
	pathIndex int
}

func (m *Move) Tick(input any, state behavior.AIState, delta time.Duration) (any, behavior.Status) {
	logger.Debug("Move - ENTER")

	if m.path == nil {
		var target mgl64.Vec3
		var ok bool

		if target, ok = input.(mgl64.Vec3); !ok {
			logger.Debug("Move - FAIL")
			return nil, behavior.FAILURE
		}

		pathManager := directory.GetDirectory().PathManager()
		position := m.Entity.Position()

		path := pathManager.FindPath(
			geometry.Point{float64(position.X()), float64(position.Y()), float64(position.Z())},
			geometry.Point{float64(target.X()), float64(target.Y()), float64(target.Z())},
		)

		if path != nil {
			m.path = path
			m.pathIndex = 1
			m.Entity.SetTarget(m.path[m.pathIndex].MglVector3())
		}
	}

	if m.path == nil {
		logger.Debug("Move - FAIL")
		return nil, behavior.FAILURE
	}

	if m.pathIndex == len(m.path) {
		logger.Debug("Move - SUCCESS")
		return nil, behavior.SUCCESS
	}

	if m.Entity.Position().Sub(m.path[m.pathIndex].MglVector3()).Len() <= 0.1 {
		m.pathIndex++
		if m.pathIndex < len(m.path) {
			m.Entity.SetTarget(m.path[m.pathIndex].MglVector3())
		}
	}

	if m.pathIndex == len(m.path) {
		logger.Debug("Move - SUCCESS")
		return nil, behavior.SUCCESS
	}

	logger.Debug("Move - RUNNING")
	return nil, behavior.RUNNING
}

func (m *Move) Reset() {
	m.path = nil
	m.pathIndex = 0
}
