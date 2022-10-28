package eventbroker

import (
	"github.com/kkevinchou/izzet/izzet/events"
)

type Observer interface {
	Observe(event events.Event)
}

type EventBroker interface {
	Broadcast(event events.Event)
	AddObserver(observer Observer, eventTypes []events.EventType)
}

type EventBrokerImpl struct {
	observations map[events.EventType][]Observer
}

func NewEventBroker() *EventBrokerImpl {
	return &EventBrokerImpl{
		observations: map[events.EventType][]Observer{},
	}
}

func (e *EventBrokerImpl) Broadcast(event events.Event) {
	for _, observer := range e.observations[event.Type()] {
		observer.Observe(event)
	}
}

func (e *EventBrokerImpl) AddObserver(observer Observer, eventTypes []events.EventType) {
	for _, eventType := range eventTypes {
		e.observations[eventType] = append(e.observations[eventType], observer)
	}
}
