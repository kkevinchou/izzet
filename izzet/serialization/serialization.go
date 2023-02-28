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
	Name     string
	ID       int
	PrefabID *int
	// Position []float64
	Position mgl64.Vec3
	Rotation mgl64.Quat
	Scale    mgl64.Vec3

	Billboard *entities.BillboardInfo
	LightInfo *entities.LightInfo
	ImageInfo *entities.ImageInfo
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
	// sEntityMap := map[string]SerializedEntity{}
	for _, entity := range s.world.Entities() {
		sEntity := SerializedEntity{
			Name:     entity.Name,
			ID:       entity.ID,
			Position: entity.LocalPosition,
			Rotation: entity.LocalRotation,
			Scale:    entity.Scale,

			ImageInfo: entity.ImageInfo,
			LightInfo: entity.LightInfo,
			Billboard: entity.Billboard,
		}

		if entity.Prefab != nil {
			id := entity.Prefab.ID
			sEntity.PrefabID = &id
		}

		serializedEntities = append(
			serializedEntities,
			sEntity,
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

func (s *Serializer) ReadIn(filepath string) error {
	f, err := os.Open(filepath)
	if err != nil {
		return err
	}

	defer f.Close()

	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	var serializedWorld SerializedWorld

	err = json.Unmarshal(bytes, &serializedWorld)
	if err != nil {
		return err
	}

	s.serializedWorld = serializedWorld
	return nil
}

func (s *Serializer) Entities() []*entities.Entity {
	dsEntities := []*entities.Entity{}
	for _, e := range s.serializedWorld.Entities {
		var dsEntity *entities.Entity
		if e.PrefabID != nil {
			dsEntity = entities.InstantiateFromPrefabStaticID(*e.PrefabID, s.world.GetPrefabByID(*e.PrefabID))
		} else {
			dsEntity = entities.InstantiateBaseEntity(e.Name, e.ID)
		}

		dsEntity.LightInfo = e.LightInfo
		dsEntity.Billboard = e.Billboard
		dsEntity.ImageInfo = e.ImageInfo

		dsEntity.LocalPosition = e.Position
		dsEntity.LocalRotation = e.Rotation
		dsEntity.Scale = e.Scale
		if dsEntity.Scale.X() == 0 && dsEntity.Scale.Y() == 0 && dsEntity.Scale.Z() == 0 {
			dsEntity.Scale = mgl64.Vec3{1, 1, 1}
		}

		dsEntities = append(dsEntities, dsEntity)
	}
	return dsEntities
}
