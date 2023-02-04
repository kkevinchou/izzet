package serialization

import (
	"encoding/json"
	"fmt"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/entities"
)

type World interface {
	Entities() []*entities.Entity
}

type SerializedWorld struct {
	Entities []SerializedEntity
}

type SerializedEntity struct {
	ID       int
	Prefab   string
	Position mgl64.Vec3
	Rotation mgl64.Quat
}

type Serializer struct {
	world World
}

func New(world World) *Serializer {
	return &Serializer{world: world}
}

func (s *Serializer) WriteOut(filepath string) {
	serializedEntities := []SerializedEntity{}
	for _, entity := range s.world.Entities() {
		serializedEntities = append(
			serializedEntities,
			SerializedEntity{
				ID:       entity.ID,
				Prefab:   entity.Prefab.Name,
				Position: entity.Position,
				Rotation: entity.Rotation,
			},
		)
	}

	serializedWorld := SerializedWorld{
		Entities: serializedEntities,
	}

	bytes, err := json.Marshal(serializedWorld)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(bytes))
}

func (s *Serializer) ReadIn(filepath string) {

}
