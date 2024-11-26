package assets

type AssetConfig struct {
	SingleEntity bool
	Name         string
	FilePath     string

	Static        bool
	Physics       bool
	ColliderType  string
	ColliderGroup string
}

type ColliderGroup string

var (
	ColliderGroupTerrain ColliderGroup = "TERRAIN"
	ColliderGroupPlayer  ColliderGroup = "PLAYER"
)

type ColliderType string

const (
	ColliderTypeNone ColliderType = "NONE"
	ColliderTypeMesh ColliderType = "MESH"
)
