package panels

type DebugSettings struct {
	DirectionalLightDir       [3]float32
	Roughness                 float32
	Metallic                  float32
	PointLightIntensity       int32
	DirectionalLightIntensity int32
	PointLightBias            float32
	MaterialOverride          bool
	EnableShadowMapping       bool
	DebugTexture              uint32 // 64 bits as we need extra bits to specify a the type of texture to IMGUI
	BloomIntensity            float32
	Exposure                  float32
	AmbientFactor             float32
	Bloom                     bool
	BloomThresholdPasses      int32
	BloomThreshold            float32
	BloomUpsamplingScale      float32
	Color                     [3]float32
	ColorIntensity            float32

	RenderSpatialPartition bool

	RenderTime float64
	FPS        float64

	FovX float32
	Near float32
	Far  float32

	TriangleDrawCount int
	DrawCount         int

	TriangleHIT                   bool
	NavMeshHSV                    bool
	NavMeshRegionIDThreshold      int32
	NavMeshDistanceFieldThreshold int32
	HSVOffset                     int32
	VoxelHighlightX               int32
	VoxelHighlightZ               int32
	VoxelHighlightDistanceField   float32
	VoxelHighlightRegionID        int
}

var DBG DebugSettings = DebugSettings{
	DirectionalLightDir:       [3]float32{-1, -1, -1},
	Roughness:                 0.55,
	Metallic:                  1.0,
	PointLightIntensity:       100,
	DirectionalLightIntensity: 5,
	PointLightBias:            1,
	MaterialOverride:          false,
	EnableShadowMapping:       false,
	BloomIntensity:            0.04,
	Exposure:                  1.0,
	AmbientFactor:             0.1,
	Bloom:                     true,
	BloomThresholdPasses:      0,
	BloomThreshold:            0.8,
	BloomUpsamplingScale:      1.0,
	Color:                     [3]float32{1, 1, 1},
	ColorIntensity:            1.0,
	RenderSpatialPartition:    false,
	FPS:                       0,

	Near: 1,
	Far:  3000,
	FovX: 105,

	TriangleDrawCount: 0,
	DrawCount:         0,

	NavMeshHSV:                    true,
	NavMeshRegionIDThreshold:      3000,
	NavMeshDistanceFieldThreshold: 23,
	HSVOffset:                     11,
	VoxelHighlightX:               0,
	VoxelHighlightZ:               0,
	VoxelHighlightDistanceField:   -1,
	VoxelHighlightRegionID:        -1,
}
