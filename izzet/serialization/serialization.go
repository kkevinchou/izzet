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
	ShapeData []*entities.ShapeData

	ChildIDs []int
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
		sEntity := SerializedEntity{
			Name:     entity.Name,
			ID:       entity.ID,
			Position: entity.LocalPosition,
			Rotation: entity.LocalRotation,
			Scale:    entity.Scale,

			ImageInfo: entity.ImageInfo,
			LightInfo: entity.LightInfo,
			Billboard: entity.Billboard,
			ShapeData: entity.ShapeData,
			ChildIDs:  []int{},
		}

		if entity.Children != nil {
			for _, child := range entity.Children {
				sEntity.ChildIDs = append(sEntity.ChildIDs, child.ID)
			}
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
	entityMap := map[int]*entities.Entity{}
	dsEntities := []*entities.Entity{}
	for _, e := range s.serializedWorld.Entities {
		var dsEntity *entities.Entity
		if e.PrefabID != nil {
			prefab := s.world.GetPrefabByID(*e.PrefabID)
			dsEntity = entities.InstantiateFromPrefabStaticID(e.ID, prefab.ModelRefs[0].Model, prefab)
		} else {
			dsEntity = entities.InstantiateBaseEntity(e.Name, e.ID)
		}

		dsEntity.LightInfo = e.LightInfo
		dsEntity.Billboard = e.Billboard
		dsEntity.ImageInfo = e.ImageInfo
		dsEntity.ShapeData = e.ShapeData

		dsEntity.LocalPosition = e.Position
		dsEntity.LocalRotation = e.Rotation
		dsEntity.Scale = e.Scale
		if dsEntity.Scale.X() == 0 && dsEntity.Scale.Y() == 0 && dsEntity.Scale.Z() == 0 {
			dsEntity.Scale = mgl64.Vec3{1, 1, 1}
		}

		entityMap[dsEntity.ID] = dsEntity
		dsEntities = append(dsEntities, dsEntity)
	}

	// set up parental relationship
	for _, e := range s.serializedWorld.Entities {
		for _, id := range e.ChildIDs {
			entityMap[e.ID].Children[id] = entityMap[id]
			entityMap[id].Parent = entityMap[e.ID]
		}
	}

	return dsEntities
}
