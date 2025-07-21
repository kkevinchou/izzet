package entities

import "github.com/go-gl/mathgl/mgl64"

type AIComponent struct {
	PatrolConfig   *PatrolConfig
	RotationConfig *RotationConfig
	TargetConfig   *TargetConfig
	PathfindConfig *PathfindConfig
	Speed          float64

	// AIVelocity mgl64.Vec3

	AttackConfig *AttackConfig
	State        AIState
}

type AIState string

var (
	AIStateIdle    AIState = "IDLE"
	AIStateAttack  AIState = "ATTACK"
	AIStatePathing AIState = "PATHING"
)

type TargetConfig struct {
	Direction mgl64.Vec3
}

type PatrolConfig struct {
	Points []mgl64.Vec3
	Index  int
}

type AttackConfig struct {
}

type PathfindingState string

var (
	PathfindingStateNoGoal  PathfindingState = ""
	PathfindingStateGoalSet PathfindingState = "GOAL_SET"
	PathfindingStatePathing PathfindingState = "PATHING"
)

type PathfindConfig struct {
	Goal       mgl64.Vec3
	Path       []mgl64.Vec3
	PolyPath   []int
	NextTarget int
	State      PathfindingState
}

type RotationConfig struct {
	Quat mgl64.Quat
}
