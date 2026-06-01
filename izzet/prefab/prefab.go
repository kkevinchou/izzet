package prefab

import "github.com/kkevinchou/izzet/izzet/assets"

type App interface {
	AssetManager() *assets.AssetManager
}
