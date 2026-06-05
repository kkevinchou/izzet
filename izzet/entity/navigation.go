package entity

import (
	"github.com/go-gl/mathgl/mgl64"
)

type NavigationComponent struct {
	Goal       mgl64.Vec3
	Path       []mgl64.Vec3
	PolyPath   []int
	NextTarget int
	State      PathfindingState
}

func (n *NavigationComponent) SetGoal(goal mgl64.Vec3) {
	n.Goal = goal
	n.State = PathfindingStateGoalSet
}

func (n *NavigationComponent) ClearGoal() {
	n.State = PathfindingStateNoGoal
}

type PathfindingState string

var (
	PathfindingStateNoGoal  PathfindingState = ""
	PathfindingStateGoalSet PathfindingState = "GOAL_SET"
	PathfindingStatePathing PathfindingState = "PATHING"
)
