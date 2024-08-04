package entities

import "github.com/go-gl/mathgl/mgl32"

type AIComponent struct {
	PatrolConfig   *PatrolConfig
	RotationConfig *RotationConfig
	TargetConfig   *TargetConfig
	PathfindConfig *PathfindConfig
	Speed          float32

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
	Direction mgl32.Vec3
}

type PatrolConfig struct {
	Points []mgl32.Vec3
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
	Goal       mgl32.Vec3
	Path       []mgl32.Vec3
	PolyPath   []int
	NextTarget int
	State      PathfindingState
}

type RotationConfig struct {
	Quat mgl32.Quat
}
