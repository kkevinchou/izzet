package animation

import "fmt"

type evalContext[T any] struct {
	game   T
	player *AnimationPlayer
}

type Condition[T any] struct {
	name string
	eval func(evalContext[T]) bool
}

func NewGameCondition[T any](name string, eval func(T) bool) Condition[T] {
	if eval == nil {
		panic(fmt.Sprintf("animation condition %q has nil evaluator", name))
	}

	return Condition[T]{
		name: name,
		eval: func(ctx evalContext[T]) bool {
			return eval(ctx.game)
		},
	}
}

func ClipCompletedCondition[T any]() Condition[T] {
	return Condition[T]{
		name: "clipCompleted",
		eval: func(ctx evalContext[T]) bool {
			return ctx.player != nil && ctx.player.NormalizedClipProgress() >= 1
		},
	}
}

func (c Condition[T]) Name() string {
	return c.name
}

func (c Condition[T]) evaluate(ctx evalContext[T]) bool {
	return c.eval(ctx)
}
