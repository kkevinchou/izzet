package animation

type MovingCondition struct {
}

func (c *MovingCondition) Name() string {
	return "MovingCondition"
}

func (c *MovingCondition) Evaluate(ctx AnimationContext) bool {
	return ctx.Moving
}

type NotMovingCondition struct {
}

func (c *NotMovingCondition) Name() string {
	return "NotMovingCondition"
}

func (c *NotMovingCondition) Evaluate(ctx AnimationContext) bool {
	return !ctx.Moving
}

type JumpTriggeredCondition struct {
}

func (c *JumpTriggeredCondition) Evaluate(ctx AnimationContext) bool {
	return ctx.JumpTriggered
}

func (c *JumpTriggeredCondition) Name() string {
	return "JumpTriggeredCondition"
}

type ClipCompletedCondition struct {
	debug bool
}

func (c *ClipCompletedCondition) Evaluate(ctx AnimationContext) bool {
	result := ctx.Player.NormalizedClipProgress() >= 1
	return result
}

func (c *ClipCompletedCondition) Name() string {
	return "ClipCompletedCondition"
}

type AirborneCondition struct {
}

func (c *AirborneCondition) Evaluate(ctx AnimationContext) bool {
	return ctx.Airborne
}

func (c *AirborneCondition) Name() string {
	return "AirborneCondition"
}

type GroundedCondition struct {
}

func (c *GroundedCondition) Evaluate(ctx AnimationContext) bool {
	return ctx.Grounded
}

func (c *GroundedCondition) Name() string {
	return "GroundedCondition"
}
