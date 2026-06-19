package event

type EventManager struct {
	PlayerJoinTopic       *Topic[PlayerJoinEvent]
	PlayerDisconnectTopic *Topic[PlayerDisconnectEvent]
	EntitySpawnTopic      *Topic[EntitySpawnEvent]
	DestroyEntityTopic    *Topic[DestroyEntityEvent]
}

func NewEventManager() *EventManager {
	return &EventManager{
		PlayerJoinTopic:       &Topic[PlayerJoinEvent]{},
		PlayerDisconnectTopic: &Topic[PlayerDisconnectEvent]{},
		EntitySpawnTopic:      &Topic[EntitySpawnEvent]{},
		DestroyEntityTopic:    &Topic[DestroyEntityEvent]{},
	}
}

type Topic[T any] struct {
	events []T
}

func (t *Topic[T]) Write(event T) {
	t.events = append(t.events, event)
}

func (t *Topic[T]) ReadFrom(cursor int) ([]T, int) {
	if cursor > len(t.events) {
		panic("what, cursor should always be at most the length of the number of events, this would imply we've cleared events which we shouldn't be doing, since we treat it as a log stream")
	}
	if cursor == len(t.events) {
		return nil, cursor
	}
	return t.events[cursor:], len(t.events)
}

type Consumer[T any] struct {
	cursor int
	topic  *Topic[T]
}

func NewConsumer[T any](topic *Topic[T]) *Consumer[T] {
	return &Consumer[T]{topic: topic}
}

func (c *Consumer[T]) ReadNewEvents() []T {
	var result []T
	result, c.cursor = c.topic.ReadFrom(c.cursor)
	return result
}
