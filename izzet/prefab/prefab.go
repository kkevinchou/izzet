package prefab

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/serialization"
)

type App interface {
	AssetManager() *assets.AssetManager
}

type Prefab struct {
	Name   string
	Handle PrefabHandle
	Entity *entity.Entity
	bytes  []byte `json:"-"`
}

type Asset struct {
	Prefab Prefab
}

type PrefabHandle struct {
	id string
}

var (
	PrefabHandleMannequin    PrefabHandle = PrefabHandle{id: "mannequin"}
	PrefabHandleVelociraptor PrefabHandle = PrefabHandle{id: "velociraptor"}
)

var PrefabRegistry map[PrefabHandle]Prefab

func InitializePrefabs(am *assets.AssetManager) {
	PrefabRegistry = map[PrefabHandle]Prefab{}
	CreateDefaultPrefabs(am)
}

func CreateDefaultPrefabs(am *assets.AssetManager) {
	player := createPlayer(am)
	velociraptor := createNPC(am, entity.EntityTypeVelociraptor)

	_ = RegisterPrefabWithHandle(PrefabHandleMannequin, "mannequin", player)
	_ = RegisterPrefabWithHandle(PrefabHandleVelociraptor, "velociraptor", velociraptor)
}

func RegisterPrefab(name string, template *entity.Entity) error {
	return RegisterPrefabWithHandle(PrefabHandle{id: name}, name, template)
}

func RegisterPrefabWithHandle(handle PrefabHandle, name string, template *entity.Entity) error {
	if template == nil {
		return fmt.Errorf("prefab [%s] template entity is nil", handle.id)
	}

	PrefabRegistry[handle] = newPrefab(handle, name, template)
	return nil
}

func Delete(handle PrefabHandle) {
	delete(PrefabRegistry, handle)
}

func Prefabs() []Prefab {
	handles := make([]PrefabHandle, 0, len(PrefabRegistry))
	for handle := range PrefabRegistry {
		handles = append(handles, handle)
	}
	sort.Slice(handles, func(i, j int) bool {
		return string(handles[i].id) < string(handles[j].id)
	})

	var prefabs []Prefab
	for _, handle := range handles {
		prefabs = append(prefabs, PrefabRegistry[handle])
	}
	return prefabs
}

func newPrefab(handle PrefabHandle, name string, e *entity.Entity) Prefab {
	bytes, err := serialization.SerializeEntity(e)
	if err != nil {
		panic(err)
	}

	return Prefab{Name: name, Entity: e, bytes: bytes, Handle: handle}
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

func (h PrefabHandle) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID string
	}{
		ID: string(h.id),
	})
}

func (h *PrefabHandle) UnmarshalJSON(data []byte) error {
	var value struct {
		ID string
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*h = PrefabHandle{id: value.ID}
	return nil
}
