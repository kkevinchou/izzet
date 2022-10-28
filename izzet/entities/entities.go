package entities

import (
	"sync"

	"github.com/kkevinchou/izzet/izzet/components"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/types"
)

var (
	nextEntityID      int = settings.EntityIDStart
	nextEntityIDMutex sync.Mutex
)

type Entity interface {
	GetID() int
	Type() types.EntityType
	GetName() string
	GetComponentContainer() *components.ComponentContainer
}

type EntityImpl struct {
	ID                 int
	entityType         types.EntityType
	Name               string
	ComponentContainer *components.ComponentContainer
}

func NewEntity(name string, entityType types.EntityType, componentContainer *components.ComponentContainer) *EntityImpl {
	entityID := GetAndIncNextEntityID()
	e := EntityImpl{
		ID:                 entityID,
		entityType:         entityType,
		Name:               name,
		ComponentContainer: componentContainer,
	}
	return &e
}

func (e *EntityImpl) GetComponentContainer() *components.ComponentContainer {
	return e.ComponentContainer
}

func (e *EntityImpl) GetName() string {
	return e.Name
}

func (e *EntityImpl) GetID() int {
	return e.ID
}

func (e *EntityImpl) Type() types.EntityType {
	return e.entityType
}

func GetAndIncNextEntityID() int {
	nextEntityIDMutex.Lock()
	defer nextEntityIDMutex.Unlock()
	nextID := nextEntityID
	nextEntityID += 1

	return nextID
}
