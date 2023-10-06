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
	Parent   int
	Children []int
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
	serializedWorld := SerializedWorld{
		Entities: s.world.Entities(),
	}

	// set up relations

	if err := writeToFile(serializedWorld, filepath); err != nil {
		panic(err)
	}
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

func (s *Serializer) ReadIn(filepath string) error {
	serializedWorld, err := readFile(filepath)
	if err != nil {
		return err
	}

	s.serializedWorld = *serializedWorld
	for _, e := range s.serializedWorld.Entities {
		e.DirtyTransformFlag = true
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
