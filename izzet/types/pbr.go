package types

type PBR struct {
	Roughness        float32
	Metallic         float32
	Diffuse          [3]float32
	DiffuseIntensity float32

	// for texture based pbr
	ColorTextureIndex       *int
	ColorTextureCoordsIndex int32
	TextureName             string
}
