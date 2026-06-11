package entity

import (
	"github.com/go-gl/mathgl/mgl64"
)

const (
	InvalidNavigationTarget int = -1
)

type NavigationComponent struct {
	Goal mgl64.Vec3
	Path []mgl64.Vec3

	PathDirty  bool
	NextTarget int
	State      PathfindingState
}

func NewNavigationComponent() *NavigationComponent {
	return &NavigationComponent{NextTarget: InvalidNavigationTarget}
}

func (n *NavigationComponent) SetGoal(goal mgl64.Vec3) {
	n.Goal = goal
	n.PathDirty = true
}

func (n *NavigationComponent) ClearGoal() {
	n.State = Idle
}

type PathfindingState string

var (
	Idle                    PathfindingState = "IDLE"
	PathfindingStatePathing PathfindingState = "PATHING"
)
