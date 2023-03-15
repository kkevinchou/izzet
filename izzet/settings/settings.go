package settings

import "github.com/inkyblackness/imgui-go/v4"

type GameMode string

var (
	// dynamic settings loaded from config
	Width      int  = 0
	Height     int  = 0
	Fullscreen bool = false
	Profile    bool = false

	ShowImguiDemoWindow   = false
	RuntimeMaxTextureSize int
)

const (
	LoggingLevel       = 1
	Seed         int64 = 1234567

	// MSPerCommandFrame is the size of the simulation step for reading input,
	// physics, etc.
	MSPerCommandFrame int = 7

	// number of rendered frames per second, separate from command frames
	FPS int = 144

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

	// perspective params
	Near float64 = 1
	Far  float64 = 3000
	FovX float64 = 105

	// shadow map properties
	ShadowmapZOffset        float64 = 400
	ShadowMapDistanceFactor float64 = .4 // proportion of view fustrum to include in shadow cuboid

	DefaultTexture string = "color_grid"
)

// styles
var (
	InActiveColorBg      imgui.Vec4 = imgui.Vec4{X: .1, Y: .1, Z: 0.1, W: 1}
	ActiveColorBg        imgui.Vec4 = imgui.Vec4{X: .3, Y: .3, Z: 0.3, W: 1}
	HoverColorBg         imgui.Vec4 = imgui.Vec4{X: .25, Y: .25, Z: 0.25, W: 1}
	InActiveColorControl imgui.Vec4 = imgui.Vec4{X: .4, Y: .4, Z: 0.4, W: 1}
	HoverColorControl    imgui.Vec4 = imgui.Vec4{X: .45, Y: .45, Z: 0.45, W: 1}
	ActiveColorControl   imgui.Vec4 = imgui.Vec4{X: .5, Y: .5, Z: 0.5, W: 1}
	HeaderColor          imgui.Vec4 = imgui.Vec4{X: 0.3, Y: 0.3, Z: 0.3, W: 1}
	HoveredHeaderColor   imgui.Vec4 = imgui.Vec4{X: 0.4, Y: 0.4, Z: 0.4, W: 1}
	TitleColor           imgui.Vec4 = imgui.Vec4{X: 0.5, Y: 0.5, Z: 0.5, W: 1}
)
