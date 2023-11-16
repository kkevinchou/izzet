package serialization

import (
	"encoding/json"
	"io"
	"os"

	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/modellibrary"
	"github.com/kkevinchou/izzet/izzet/world"
)

type App interface {
	ModelLibrary() *modellibrary.ModelLibrary
}

type GameWorld interface {
	Entities() []*entities.Entity
}

type Relation struct {
	Parent int
	Child  int
}

type WorldIR struct {
	Entities  []*entities.Entity
	Relations []Relation
}

type Serializer struct {
	app App
}

func New(app App) *Serializer {
	return &Serializer{app: app}
}

func (s *Serializer) WriteToFile(world GameWorld, filepath string) error {
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}

	defer f.Close()

	err = s.Write(world, f)
	if err != nil {
		return err
	}

	return nil
}

func (s *Serializer) Write(world GameWorld, writer io.Writer) error {
	entities := world.Entities()

	worldIR := WorldIR{
		Entities: entities,
	}

	for _, entity := range entities {
		if entity.Parent != nil {
			worldIR.Relations = append(worldIR.Relations, Relation{Parent: entity.Parent.ID, Child: entity.ID})
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

func (s *Serializer) ReadFromFile(filepath string) (*world.GameWorld, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	gameWorld, err := s.Read(f)
	if err != nil {
		return nil, err
	}

	return gameWorld, err
}

func (s *Serializer) Read(reader io.Reader) (*world.GameWorld, error) {
	// bytes, err := io.ReadAll(reader)
	// if err != nil {
	// 	return nil, err
	// }

	var worldIR WorldIR
	decoder := json.NewDecoder(reader)
	err := decoder.Decode(&worldIR)
	// err = json.Unmarshal(bytes, &worldIR)
	if err != nil {
		return nil, err
	}

	entityMap := map[int]*entities.Entity{}
	for _, entity := range worldIR.Entities {
		entityMap[entity.ID] = entity
		InitDeserializedEntity(entity, s.app.ModelLibrary())
	}

	// rebuild relations
	for _, relation := range worldIR.Relations {
		parent := entityMap[relation.Parent]
		child := entityMap[relation.Child]
		if len(parent.Children) == 0 {
			parent.Children = make(map[int]*entities.Entity)
		}
		parent.Children[relation.Child] = child
		child.Parent = parent
	}

	return world.New(entityMap), nil
}
