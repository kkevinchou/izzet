package gltf

const (
	TextureCoordStyleOpenGL = 1
)

type TextureCoordStyle int

type ParseConfig struct {
	TextureCoordStyle TextureCoordStyle
}
