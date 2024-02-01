package settings

type GameMode string

// var (
// 	// dynamic settings loaded from config
// 	Width      int  = 0
// 	Height     int  = 0
// 	Fullscreen bool = false
// 	Profile    bool = false

// 	RuntimeMaxTextureSize int
// )

type Config struct {
	Width         int
	Height        int
	Fullscreen    bool
	Profile       bool
	ServerAddress string `json:"server_address"`
}

func NewConfig() Config {
	return Config{
		Width:         0,
		Height:        0,
		Fullscreen:    false,
		Profile:       false,
		ServerAddress: "localhost:7878",
	}
}

const (
	LoggingLevel       = 1
	Seed         int64 = 1234567

	// MSPerCommandFrame is the size of the simulation step for reading input,
	// physics, etc.
	MSPerCommandFrame    int = 16
	MSPerGameStateUpdate int = 100

	// number of rendered frames per second, separate from command frames
	FPS int = 300

	// the maximum number of command frames to execute in a single loop to prevent the spiral of death
	MaxCommandFramesPerLoop int = 3

	// Animation
	MaxAnimationJointWeights = 4

	maxUInt32           uint32 = ^uint32(0)
	EmptyColorPickingID uint32 = maxUInt32

	// this number should like up with MAX_LIGHTS in the fragment shader
	MaxLightCount int = 10

	DepthCubeMapWidth  float32 = 4096
	DepthCubeMapHeight float32 = 4096
	DepthCubeMapNear   float64 = 1
	DepthCubeMapFar    float64 = 800

	// shadow map properties
	ShadowmapZOffset float64 = 600 // Z offset relative to the light's view. if this is too small, objects behind a camera may fail to cast shadows
	// ShadowMapDistanceFactor float64 = .4  // proportion of view fustrum to include in shadow cuboid
	ShadowMapDistanceFactor float64 = 1 // proportion of view fustrum to include in shadow cuboid

	DefaultTexture string = "color_grid"

	GizmoAxisThickness  float64 = 0.08
	GizmoDistanceFactor float64 = 8

	MaxCommandFrameBufferSize int = 100000
	MaxStateBufferSize        int = 100

	FirstPersonCamera  bool    = false
	CameraEntityOffset float64 = 100
	ProjectsDirectory  string  = ".project/"
	DefaultProject     string  = "default"
)

// styles
var (
	FontSize float32 = 20
)
