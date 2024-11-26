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
