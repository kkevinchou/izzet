package serialization

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/prefabs"
)

type World interface {
	Entities() []*entities.Entity
	GetPrefabByID(id int) *prefabs.Prefab
}

type SerializedWorld struct {
	Entities []SerializedEntity
}

type SerializedEntity struct {
	ID       int
	PrefabID int
	// Position []float64
	Position mgl64.Vec3
	Rotation mgl64.Quat
}

type Serializer struct {
	world           World
	serializedWorld SerializedWorld
}

func New(world World) *Serializer {
	return &Serializer{world: world}
}

func (s *Serializer) WriteOut(filepath string) {
	serializedEntities := []SerializedEntity{}
	for _, entity := range s.world.Entities() {
		position := entity.Position
		serializedEntities = append(
			serializedEntities,
			SerializedEntity{
				ID:       entity.ID,
				PrefabID: entity.Prefab.ID,
				// Position: []float64{position.X(), position.Y(), position.Z(),
				Position: position,
				Rotation: entity.Rotation,
			},
		)
	}

	serializedWorld := SerializedWorld{
		Entities: serializedEntities,
	}

	bytes, err := json.MarshalIndent(serializedWorld, "", "    ")
	if err != nil {
		panic(err)
	}

	f, err := os.Create(filepath)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	_, err = f.Write(bytes)
	if err != nil {
		panic(err)
	}
}

func (s *Serializer) ReadIn(filepath string) {
	f, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	var serializedWorld SerializedWorld

	err = json.Unmarshal(bytes, &serializedWorld)
	if err != nil {
		panic(err)
	}

	s.serializedWorld = serializedWorld
}

func (s *Serializer) Entities() []*entities.Entity {
	dsEntities := []*entities.Entity{}
	for _, e := range s.serializedWorld.Entities {
		dsEntity := entities.InstantiateFromPrefabStaticID(e.ID, s.world.GetPrefabByID(e.PrefabID))
		dsEntity.Position = e.Position
		dsEntity.Rotation = e.Rotation
		dsEntities = append(dsEntities, dsEntity)
	}
	return dsEntities
}
