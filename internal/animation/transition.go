package animation

type transition[T any] interface {
	SourceState() *AnimationState
	NextState() *AnimationState
	Evaluate(ctx evalContext[T]) bool
}

type transitionImpl[T any] struct {
	name       string
	source     *AnimationState
	target     *AnimationState
	conditions []Condition[T]
}

func (t *transitionImpl[T]) AddCondition(c Condition[T]) {
	t.conditions = append(t.conditions, c)
}

func (t *transitionImpl[T]) SourceState() *AnimationState {
	return t.source
}

func (t *transitionImpl[T]) NextState() *AnimationState {
	return t.target
}

func (t *transitionImpl[T]) Evaluate(ctx evalContext[T]) bool {
	transition := true

	for _, c := range t.conditions {
		if !c.evaluate(ctx) {
			transition = false
			break
		}
	}

	return transition
}
