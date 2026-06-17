package runtimeconfig

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/settings"
)

type RuntimeConfig struct {
	CameraPosition mgl64.Vec3
	CameraRotation mgl64.Quat

	DirectionalLightDir             [3]float32
	Roughness                       float32
	Metallic                        float32
	PointLightBias                  float32
	ShadowMapMinBias                float32
	ShadowMapAngleBiasRate          float32
	EnableShadowMapping             bool
	ShadowFarDistance               float32
	ShadowNearDistance              float32
	ShadowCascadeBlendFactor        float32
	ShadowMapMaxCascadeIndex        int32
	ShadowSpatialPartitionNearPlane float32
	BloomIntensity                  float32
	Exposure                        float32
	AmbientFactor                   float32
	SpecularFactor                  float32
	Bloom                           bool
	BloomThresholdPasses            int32
	BloomThreshold                  float32
	BloomUpsamplingScale            float32
	Color                           [3]float32

	ShowImguiDemo          bool
	ShowTextureViewer      bool
	ShowMaterialEditor     bool
	DebugTexture           uint32  // 64 bits as we need extra bits to specify a the type of texture to IMGUI
	DebugAspectRatio       float64 // the aspect ratio of the debug texture
	TextureArrayDebugLayer int32

	RenderSpatialPartition bool

	// RenderTime       float64
	// FPS              float64
	// CommandFrameTime float64

	FovX float32
	Near float32
	Far  float32

	FogStart   int32
	FogEnd     int32
	FogDensity int32

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
	SnapSize            float64
	RotationSnapSize    int32
	RotationSensitivity int32

	// Other
	UIEnabled                       bool
	SimplifyMeshIterations          int32
	ShowSelectionBoundingBox        bool
	LockRenderingToCommandFrameRate bool

	// Navigation Mesh
	NavigationMeshIterations           int32
	NavigationMeshWalkableHeight       float32
	NavigationMeshClimbableHeight      float32
	NavigationMeshMinRegionArea        int32
	NavigationMeshAgentRadius          float32
	NavigationMeshCellSize             float32
	NavigationMeshCellHeight           float32
	NavigationmeshMaxError             float32
	NavigationmeshMaxEdgeLength        float32
	NavigationmeshSampleDist           float32
	NavigationMeshFilterLedgeSpans     bool
	NavigationMeshFilterLowHeightSpans bool

	NavigationMeshStart      int32
	NavigationMeshStartPoint mgl64.Vec3
	NavigationMeshGoal       int32
	NavigationMeshGoalPoint  mgl64.Vec3

	ShadowmapZOffset float32

	DebugBlob1         string
	DebugBlob1IntMap   map[int]bool
	DebugBlob1IntSlice []int
	DebugBlob2         string
	DebugBlob2IntMap   map[int]bool
	DebugBlob2IntSlice []int

	SkyboxTopColor    [3]float32
	SkyboxBottomColor [3]float32
	SkyboxMixValue    float32

	ActiveCloudTextureIndex        int
	ActiveCloudTextureChannelIndex int
	CloudTextures                  [2]CloudTexture

	// post processing
	KuwaharaFilter bool

	// SSAO
	SSAORadius float32
	SSAOBias   float32

	BatchRenderingEnabled bool

	EnableSSAO bool

	ShowHUD bool

	TestPosition       mgl32.Vec3
	TestAngle          int32
	TestObjectRotation int32

	TestAlbedo    [3]float32
	TestMetallic  float32
	TestRoughness float32
	TestAO        float32

	EnableAntialiasing bool
}

type CloudTextureChannel struct {
	// Noise - Cloud Texture
	NoiseZ                           float32
	CellWidth, CellHeight, CellDepth int32
}

type CloudTexture struct {
	Channels                                        [4]CloudTextureChannel
	TextureWidth, TextureHeight                     int32
	WorkGroupWidth, WorkGroupHeight, WorkGroupDepth int32

	// rendering
	VAO           uint32
	WorleyTexture uint32
	FBO           uint32
	RenderTexture uint32
	ColorChannel  string
}

func DefaultRuntimeConfig() *RuntimeConfig {
	config := &RuntimeConfig{
		DirectionalLightDir:      [3]float32{-1, -1, -1},
		Roughness:                0.55,
		Metallic:                 0,
		PointLightBias:           0.3,
		ShadowMapMinBias:         0,
		ShadowMapAngleBiasRate:   0,
		EnableShadowMapping:      true,
		ShadowFarDistance:        200,
		ShadowNearDistance:       0.01,
		ShadowCascadeBlendFactor: 0.70,
		ShadowMapMaxCascadeIndex: int32(settings.NumShadowMapCascades - 1),
		ShadowmapZOffset:         250,
		BloomIntensity:           0.04,
		Exposure:                 1.0,
		AmbientFactor:            0.1,
		SpecularFactor:           1.0,
		Bloom:                    true,
		BloomThresholdPasses:     1,
		BloomThreshold:           0.8,
		BloomUpsamplingScale:     1.0,
		Color:                    [3]float32{1, 1, 1},
		RenderSpatialPartition:   false,

		Near: 0.1,
		Far:  500,
		FovX: float32(settings.DefaultFOVX),

		FogStart:   200,
		FogEnd:     1000,
		FogDensity: 1,

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

		SnapSize:            0.2,
		RotationSnapSize:    20,
		RotationSensitivity: 200,

		ShowSelectionBoundingBox: false,
		ShowColliders:            false,
		ShowTextureViewer:        false,

		NavigationMeshIterations:           2500,
		NavigationMeshWalkableHeight:       float32(settings.EntityCapsuleColliderLength + (2 * settings.EntityCapsuleColliderRadius)),
		NavigationMeshClimbableHeight:      0.3,
		NavigationMeshMinRegionArea:        4,
		NavigationMeshAgentRadius:          float32(settings.EntityCapsuleColliderRadius),
		NavigationMeshCellSize:             0.1,
		NavigationMeshCellHeight:           0.1,
		NavigationmeshMaxError:             1,
		NavigationmeshMaxEdgeLength:        500,
		NavigationmeshSampleDist:           1,
		NavigationMeshFilterLedgeSpans:     true,
		NavigationMeshFilterLowHeightSpans: true,

		NavigationMeshStart: 0,
		NavigationMeshGoal:  1,

		SkyboxTopColor:    [3]float32{0.02, 0.02, 0.32},
		SkyboxBottomColor: [3]float32{0.11, 0.93, 0.87},
		SkyboxMixValue:    0.4,

		CloudTextures: [2]CloudTexture{
			{
				Channels: [4]CloudTextureChannel{
					{
						NoiseZ:     0,
						CellWidth:  10,
						CellHeight: 10,
						CellDepth:  10,
					},
					{
						NoiseZ:     0,
						CellWidth:  10,
						CellHeight: 10,
						CellDepth:  10,
					},
					{
						NoiseZ:     0,
						CellWidth:  10,
						CellHeight: 10,
						CellDepth:  10,
					},
					{
						NoiseZ:     0,
						CellWidth:  10,
						CellHeight: 10,
						CellDepth:  10,
					},
				},
				WorkGroupWidth:  128,
				WorkGroupHeight: 128,
				WorkGroupDepth:  128,
			},
			{
				Channels: [4]CloudTextureChannel{
					{
						NoiseZ:     0,
						CellWidth:  10,
						CellHeight: 10,
						CellDepth:  10,
					},
					{
						NoiseZ:     0,
						CellWidth:  10,
						CellHeight: 10,
						CellDepth:  10,
					},
					{
						NoiseZ:     0,
						CellWidth:  10,
						CellHeight: 10,
						CellDepth:  10,
					},
					{
						NoiseZ:     0,
						CellWidth:  10,
						CellHeight: 10,
						CellDepth:  10,
					},
				},
				WorkGroupWidth:  128,
				WorkGroupHeight: 128,
				WorkGroupDepth:  128,
			},
		},
		EnableSSAO:         true,
		SSAORadius:         5,
		SSAOBias:           0.025,
		TestPosition:       mgl32.Vec3{0, 4, 6},
		TestAngle:          -35,
		TestObjectRotation: 35,
		TestAlbedo:         [3]float32{0.5, 0.5, 0.5},
		TestMetallic:       0,
		TestRoughness:      1,
		TestAO:             1,
		EnableAntialiasing: true,
	}

	return config
}
