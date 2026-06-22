package prefab

import (
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/serialization"
)

type App interface {
	AssetManager() *assets.AssetManager
}

type Prefab struct {
	Entity *entity.Entity
	bytes  []byte `json:"-"`
}

var PrefabRegistry []Prefab

func CreateDefaultPrefabs(app App) {
	player := CreatePlayer(app)
	velociraptor := CreateNPC(app, entity.EntityTypeVelociraptor)
	// parasaurolophus := CreateNPC(app, entity.EntityTypeParasaurolophus)
	PrefabRegistry = append(PrefabRegistry,
		New(player),
		New(velociraptor),
	)
}

func New(e *entity.Entity) Prefab {
	// p := Prefab{Entity: e},
	bytes, err := serialization.SerializeEntity(e)
	if err != nil {
		panic(err)
	}

	return Prefab{Entity: e, bytes: bytes}
}

func Instantiate(p Prefab, am *assets.AssetManager) *entity.Entity {
	e, err := serialization.DeserializeEntity(p.bytes, am)
	if err != nil {
		panic(err)
	}
	id := entity.GetNextIDAndAdvance()
	e.ID = id
	return e
}
