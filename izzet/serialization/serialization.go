package serialization

import (
	"encoding/json"
	"io"
	"os"

	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/prefabs"
)

type World interface {
	Entities() []*entities.Entity
	GetPrefabByID(id int) *prefabs.Prefab
}

type Relation struct {
	Parent int
	Child  int
}

type SerializedWorld struct {
	Entities  []*entities.Entity
	Relations []Relation
}

type Serializer struct {
	world           World
	serializedWorld SerializedWorld
}

func New(world World) *Serializer {
	return &Serializer{world: world}
}

func (s *Serializer) WriteOut(filepath string) {
	entities := s.world.Entities()

	serializedWorld := SerializedWorld{
		Entities: entities,
	}

	for _, e := range entities {
		if e.Parent != nil {
			serializedWorld.Relations = append(serializedWorld.Relations, Relation{Parent: e.Parent.ID, Child: e.ID})
		}
	}

	if err := writeToFile(serializedWorld, filepath); err != nil {
		panic(err)
	}
}

func (s *Serializer) ReadIn(filepath string) error {
	serializedWorld, err := readFile(filepath)
	if err != nil {
		return err
	}

	s.serializedWorld = *serializedWorld
	entityMap := map[int]*entities.Entity{}
	for _, e := range s.serializedWorld.Entities {
		e.DirtyTransformFlag = true
		entityMap[e.ID] = e
	}

	for _, relation := range s.serializedWorld.Relations {
		parent := entityMap[relation.Parent]
		child := entityMap[relation.Child]
		if len(parent.Children) == 0 {
			parent.Children = make(map[int]*entities.Entity)
		}
		parent.Children[relation.Child] = child
		child.Parent = parent
	}
	return nil
}

func writeToFile(world SerializedWorld, filepath string) error {
	bytes, err := json.MarshalIndent(world, "", "    ")
	if err != nil {
		return err
	}

	f, err := os.Create(filepath)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}

func readFile(filepath string) (*SerializedWorld, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	bytes, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var serializedWorld SerializedWorld
	err = json.Unmarshal(bytes, &serializedWorld)
	if err != nil {
		return nil, err
	}

	return &serializedWorld, err
}

func (s *Serializer) Entities() []*entities.Entity {
	return s.serializedWorld.Entities
}
