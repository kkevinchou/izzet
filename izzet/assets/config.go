package assets

type AssetConfig struct {
	Name     string
	FilePath string

	Static        bool
	Physics       bool
	ColliderType  string
	ColliderGroup string
}
