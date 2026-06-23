package prefab

import (
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
	Entity *entity.Entity
	bytes  []byte `json:"-"`
}

type Asset struct {
	Name   string
	Prefab Prefab
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

	_ = RegisterTemplate(string(PrefabHandleMannequin), player)
	_ = RegisterTemplate(string(PrefabHandleVelociraptor), velociraptor)
}

func RegisterTemplate(name string, template *entity.Entity) error {
	if name == "" {
		return fmt.Errorf("prefab name is required")
	}
	if template == nil {
		return fmt.Errorf("prefab [%s] template entity is nil", name)
	}

	handle := PrefabHandle(name)
	if _, ok := PrefabRegistry[handle]; ok {
		return fmt.Errorf("prefab [%s] already exists", name)
	}

	PrefabRegistry[handle] = newPrefab(template)
	return nil
}

func Delete(handle PrefabHandle) {
	delete(PrefabRegistry, handle)
}

func SaveAssets() []Asset {
	handles := make([]PrefabHandle, 0, len(PrefabRegistry))
	for handle := range PrefabRegistry {
		handles = append(handles, handle)
	}
	sort.Slice(handles, func(i, j int) bool {
		return string(handles[i]) < string(handles[j])
	})

	var assets []Asset
	for _, handle := range handles {
		assets = append(assets, Asset{
			Name:   string(handle),
			Prefab: PrefabRegistry[handle],
		})
	}
	return assets
}

func LoadAssets(app App, assets []Asset) error {
	PrefabRegistry = map[PrefabHandle]Prefab{}

	CreateDefaultPrefabs(app)

	for _, asset := range assets {
		if asset.Name == "" || asset.Prefab.Entity == nil {
			continue
		}
		if err := RegisterTemplate(asset.Name, asset.Prefab.Entity); err != nil {
			return err
		}
	}
	return nil
}

func newPrefab(e *entity.Entity) Prefab {
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
