package prefab

import (
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/serialization"
)

type App interface {
	AssetManager() *assets.AssetManager
}

type Prefab struct {
	ID     PrefabID
	Name   string
	Entity *entity.Entity
	bytes  []byte `json:"-"`
}

type Asset struct {
	Prefab Prefab
}

type PrefabID string

var (
	PrefabIDMannequin    PrefabID = "mannequin"
	PrefabIDVelociraptor PrefabID = "velociraptor"
)

var PrefabRegistry map[PrefabID]Prefab

func InitializePrefabs(am *assets.AssetManager) {
	PrefabRegistry = map[PrefabID]Prefab{}
	CreateDefaultPrefabs(am)
}

func CreateDefaultPrefabs(am *assets.AssetManager) {
	player := createPlayer(am)
	velociraptor := createNPC(am, entity.EntityTypeVelociraptor)

	_ = RegisterPrefabWithID(PrefabIDMannequin, "mannequin", player)
	_ = RegisterPrefabWithID(PrefabIDVelociraptor, "velociraptor", velociraptor)
}

func RegisterPrefab(name string, template *entity.Entity) error {
	return RegisterPrefabWithID(PrefabID(uuid.NewString()), name, template)
}

func RegisterPrefabWithID(id PrefabID, name string, template *entity.Entity) error {
	if id == "" {
		return fmt.Errorf("prefab id is required")
	}
	if template == nil {
		return fmt.Errorf("prefab [%s] template entity is nil", id)
	}

	PrefabRegistry[id] = newPrefab(id, name, template)
	return nil
}

func Delete(id PrefabID) {
	delete(PrefabRegistry, id)
}

func Prefabs() []Prefab {
	ids := make([]PrefabID, 0, len(PrefabRegistry))
	for id := range PrefabRegistry {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		return string(ids[i]) < string(ids[j])
	})

	var prefabs []Prefab
	for _, id := range ids {
		prefabs = append(prefabs, PrefabRegistry[id])
	}
	return prefabs
}

func newPrefab(id PrefabID, name string, e *entity.Entity) Prefab {
	bytes, err := serialization.SerializeEntity(e)
	if err != nil {
		panic(err)
	}

	return Prefab{ID: id, Name: name, Entity: e, bytes: bytes}
}

func Instantiate(id PrefabID, am *assets.AssetManager) *entity.Entity {
	p := PrefabRegistry[id]
	e, err := serialization.DeserializeEntity(p.bytes, am)
	if err != nil {
		panic(err)
	}
	entityID := entity.GetNextIDAndAdvance()
	e.ID = entityID
	return e
}
