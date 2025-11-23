package serialization

import (
	"encoding/json"
	"io"
	"os"

	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/world"
)

type GameWorld interface {
	Entities() []*entity.Entity
}

type Relation struct {
	Parent int
	Child  int
}

type WorldIR struct {
	Entities  []*entity.Entity
	Relations []Relation
}

func WriteToFile(world GameWorld, filepath string) error {
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}

	defer f.Close()

	err = Write(world, f)
	if err != nil {
		return err
	}

	return nil
}

func Write(world GameWorld, writer io.Writer) error {
	entities := world.Entities()

	worldIR := WorldIR{
		Entities: entities,
	}

	for _, e := range entities {
		if e.Parent != nil {
			worldIR.Relations = append(worldIR.Relations, Relation{Parent: e.Parent.ID, Child: e.ID})
		}
	}

	bytes, err := json.MarshalIndent(worldIR, "", "    ")
	if err != nil {
		return err
	}

	_, err = writer.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}

func ReadFromFile(filepath string, am *assets.AssetManager) (*world.GameWorld, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	gameWorld, err := Read(f, am)
	if err != nil {
		return nil, err
	}

	return gameWorld, err
}

func Read(reader io.Reader, am *assets.AssetManager) (*world.GameWorld, error) {
	var worldIR WorldIR
	decoder := json.NewDecoder(reader)
	err := decoder.Decode(&worldIR)
	if err != nil {
		return nil, err
	}

	entityMap := map[int]*entity.Entity{}
	for _, e := range worldIR.Entities {
		e.Children = make(map[int]*entity.Entity)
		entityMap[e.ID] = e
	}

	// rebuild relations
	for _, relation := range worldIR.Relations {
		parent := entityMap[relation.Parent]
		child := entityMap[relation.Child]
		parent.Children[relation.Child] = child
		child.Parent = parent
	}

	for _, e := range entityMap {
		initDeserializedEntity(e, am)
	}

	return world.NewWithEntities(entityMap), nil
}
