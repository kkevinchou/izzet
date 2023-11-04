package app

import "github.com/go-gl/mathgl/mgl64"

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
	ColorIntensity       float32

	ShowDebugTexture bool
	DebugTexture     uint32 // 64 bits as we need extra bits to specify a the type of texture to IMGUI

	EnableSpatialPartition bool
	RenderSpatialPartition bool

	RenderTime             float64
	FPS                    float64
	CommandFrameTime       float64
	CommandFramesPerRender int

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
	RenderColliders                bool
	PartitionEntityCount           int32
	PhysicsCollisionCheckCount     int32
	PhysicsConfirmedCollisionCount int32

	UIEnabled bool
}
