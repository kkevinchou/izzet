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

type PrefabHandle string

var (
	PrefabHandleMannequin    PrefabHandle = "mannequin"
	PrefabHandleVelociraptor PrefabHandle = "velociraptor"
)

var PrefabRegistry map[PrefabHandle]Prefab

func init() {
	PrefabRegistry = map[PrefabHandle]Prefab{}
}

func CreateDefaultPrefabs(app App) {
	player := createPlayer(app)
	velociraptor := createNPC(app, entity.EntityTypeVelociraptor)

	PrefabRegistry[PrefabHandleMannequin] = New(player)
	PrefabRegistry[PrefabHandleVelociraptor] = New(velociraptor)
}

func New(e *entity.Entity) Prefab {
	bytes, err := serialization.SerializeEntity(e)
	if err != nil {
		panic(err)
	}

	return Prefab{Entity: e, bytes: bytes}
}

func Instantiate(handle PrefabHandle, am *assets.AssetManager) *entity.Entity {
	p := PrefabRegistry[handle]
	e, err := serialization.DeserializeEntity(p.bytes, am)
	if err != nil {
		panic(err)
	}
	id := entity.GetNextIDAndAdvance()
	e.ID = id
	return e
}
