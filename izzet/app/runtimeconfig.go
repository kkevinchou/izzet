package app

import (
	"github.com/go-gl/mathgl/mgl64"
)

type RuntimeConfig struct {
	CameraPosition mgl64.Vec3
	CameraRotation mgl64.Quat

	DirectionalLightDir  [3]float32
	Roughness            float32
	Metallic             float32
	PointLightBias       float32
	MaterialOverride     bool
	EnableShadowMapping  bool
	ShadowFarFactor      float32
	SPNearPlaneOffset    float32
	BloomIntensity       float32
	Exposure             float32
	AmbientFactor        float32
	Bloom                bool
	BloomThresholdPasses int32
	BloomThreshold       float32
	BloomUpsamplingScale float32
	Color                [3]float32

	ShowDebugTexture bool
	DebugTexture     uint32 // 64 bits as we need extra bits to specify a the type of texture to IMGUI

	EnableSpatialPartition bool
	RenderSpatialPartition bool

	RenderTime       float64
	FPS              float64
	CommandFrameTime float64

	FovX float32
	Near float32
	Far  float32

	FogStart   int32
	FogEnd     int32
	FogDensity int32
	FogEnabled bool

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

	// Physics / Collision
	ShowColliders                  bool
	PartitionEntityCount           int32
	PhysicsCollisionCheckCount     int32
	PhysicsConfirmedCollisionCount int32

	// Editing
	SnapSize            int32
	RotationSnapSize    int32
	RotationSensitivity int32

	// Other
	UIEnabled                       bool
	SimplifyMeshIterations          int32
	ShowSelectionBoundingBox        bool
	LockRenderingToCommandFrameRate bool
}

func DefaultRuntimeConfig() RuntimeConfig {
	return RuntimeConfig{
		DirectionalLightDir:    [3]float32{-1, -1, -1},
		Roughness:              0.55,
		Metallic:               1.0,
		PointLightBias:         1,
		MaterialOverride:       false,
		EnableShadowMapping:    true,
		ShadowFarFactor:        1,
		SPNearPlaneOffset:      300,
		BloomIntensity:         0.04,
		Exposure:               1.0,
		AmbientFactor:          0.1,
		Bloom:                  true,
		BloomThresholdPasses:   1,
		BloomThreshold:         0.8,
		BloomUpsamplingScale:   1.0,
		Color:                  [3]float32{1, 1, 1},
		RenderSpatialPartition: false,
		EnableSpatialPartition: true,
		FPS:                    0,

		Near: 1,
		Far:  3000,
		FovX: 105,

		FogStart:   200,
		FogEnd:     1000,
		FogDensity: 1,
		FogEnabled: true,

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

		UIEnabled: true,

		SnapSize:            1,
		RotationSnapSize:    20,
		RotationSensitivity: 200,

		ShowSelectionBoundingBox: true,
		ShowColliders:            false,
	}
}
