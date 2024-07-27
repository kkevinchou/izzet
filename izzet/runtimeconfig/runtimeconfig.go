package runtimeconfig

import (
	"github.com/go-gl/mathgl/mgl64"
)

type RuntimeConfig struct {
	CameraPosition mgl64.Vec3
	CameraRotation mgl64.Quat

	DirectionalLightDir             [3]float32
	Roughness                       float32
	Metallic                        float32
	PointLightBias                  float32
	MaterialOverride                bool
	EnableShadowMapping             bool
	ShadowFarDistance               float32
	ShadowSpatialPartitionNearPlane float32
	BloomIntensity                  float32
	Exposure                        float32
	AmbientFactor                   float32
	Bloom                           bool
	BloomThresholdPasses            int32
	BloomThreshold                  float32
	BloomUpsamplingScale            float32
	Color                           [3]float32

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

	// Navigation Mesh
	NavigationMeshIterations           int32
	NavigationMeshWalkableHeight       int32
	NavigationMeshClimbableHeight      int32
	NavigationMeshMinRegionArea        int32
	NavigationMeshAgentRadius          float32
	NavigationMeshCellSize             float32
	NavigationMeshCellHeight           float32
	NavigationmeshMaxError             float32
	NavigationmeshMaxEdgeLength        float32
	NavigationmeshSampleDist           float32
	NavigationMeshFilterLedgeSpans     bool
	NavigationMeshFilterLowHeightSpans bool

	NavigationMeshStart int32
	NavigationMeshGoal  int32

	ShadowmapZOffset float32

	DebugBlob1         string
	DebugBlob1IntMap   map[int]bool
	DebugBlob1IntSlice []int
	DebugBlob2         string
	DebugBlob2IntMap   map[int]bool
	DebugBlob2IntSlice []int
}

func DefaultRuntimeConfig() RuntimeConfig {
	return RuntimeConfig{
		DirectionalLightDir:             [3]float32{-1, -1, -1},
		Roughness:                       0.55,
		Metallic:                        0,
		PointLightBias:                  0.5,
		MaterialOverride:                false,
		EnableShadowMapping:             true,
		ShadowFarDistance:               200,
		ShadowSpatialPartitionNearPlane: 1000,
		ShadowmapZOffset:                1000,
		BloomIntensity:                  0.04,
		Exposure:                        1.0,
		AmbientFactor:                   0.1,
		Bloom:                           true,
		BloomThresholdPasses:            1,
		BloomThreshold:                  0.8,
		BloomUpsamplingScale:            1.0,
		Color:                           [3]float32{1, 1, 1},
		RenderSpatialPartition:          false,
		EnableSpatialPartition:          true,
		FPS:                             0,

		Near: 1,
		Far:  500,
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

		NavigationMeshIterations:           500,
		NavigationMeshWalkableHeight:       4,
		NavigationMeshClimbableHeight:      3,
		NavigationMeshMinRegionArea:        4,
		NavigationMeshAgentRadius:          0.6,
		NavigationMeshCellSize:             0.5,
		NavigationMeshCellHeight:           0.5,
		NavigationmeshMaxError:             1,
		NavigationmeshMaxEdgeLength:        6,
		NavigationmeshSampleDist:           1,
		NavigationMeshFilterLedgeSpans:     true,
		NavigationMeshFilterLowHeightSpans: true,

		NavigationMeshStart: 0,
		NavigationMeshGoal:  1,
	}
}
