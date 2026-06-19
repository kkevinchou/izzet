package events

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

func (em *EventManager) Clear() {
	em.PlayerJoinTopic.Clear()
	em.PlayerDisconnectTopic.Clear()
	em.EntitySpawnTopic.Clear()
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

// right now all topics are an append only stream that consumers can read any number of events
// this means we preserve all history. i might revisit this decision later since this will eventually
// run out of memory. but it's nice in that we could theoretically switch to a ring buffer in the future
// and avoid resizing an array. however, i don't want to use a ring buffer yet, maybe later

// also this is a kinda cool setup because it allows systems that don't run every frame (like replication)
// to catch up on events that have piled up between frames.
func (t *Topic[T]) Clear() {
	t.events = nil
	panic("first time calling clear")
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
