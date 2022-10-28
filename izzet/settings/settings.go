package settings

import (
	"time"

	"github.com/go-gl/mathgl/mgl64"
)

type GameMode string

var (
	CurrentGameMode     GameMode = GameModeUndefined
	CameraStartPosition          = mgl64.Vec3{0, 10, 30}
	CameraStartView              = mgl64.Vec2{0, 0}
	ListenAddress       string   = "localhost"

	AccelerationDueToGravity = mgl64.Vec3{0, -gravity, 0}

	// dynamic settings loaded from config
	Host       string = "localhost"
	Port       int    = 8080
	Width      int    = 0
	Height     int    = 0
	Fullscreen bool   = false

	// dynamic settings configurable from the console
	DebugRenderCollisionVolume  = false
	DebugRenderSpatialPartition = false

	ShowImguiDemoWindow   = false
	RuntimeMaxTextureSize int
)

const (
	LoggingLevel          = 1
	Seed           int64  = 1234567
	ConnectionType string = "tcp"

	GameModeUndefined GameMode = "UNDEFINED"
	GameModeClient    GameMode = "CLIENT"
	GameModeServer    GameMode = "SERVER"

	ServerID      int = 69
	ClientIDStart int = 70000
	EntityIDStart int = 80000

	PProfEnabled    bool = false
	PProfClientPort int  = 6060
	PProfServerPort int  = 6061

	// MSPerCommandFrame is the size of the simulation step for reading input,
	// physics, etc.
	MSPerCommandFrame int = 16
	FPS               int = 60

	// MaxInputBufferCommandFrames controls how many command frames we buffer client inputs for. See the
	// InputBuffer struct definition for a more detailed description.

	// This buffer size should ideally be able to fully contain and fully sim a singular
	// player action. Otherwise, there's an edge case where a player starts an action and
	// the message takes more than a command frame of time to reach the server, causing the
	// player action to apply one or more frames late which causes a client misprediction.
	// If we were in the realm of the command buffer, we would have been able to place the
	// message in the buffer and "push" it forward and execute on the correct frame.

	// The max command frames for the buffer is currently static but is ideally dynamically
	// able to resize depending on the quality of the player's internet connection to the server.
	// The faster the connection, the smaller the buffer needs to be may be something we want to be
	// able to dyanmically adjust based on player latency. The larger their latency, the larger the
	// buffer

	// This is potentially overkill to avoiding absolutely no mispredictions on the client.
	// The drawback of an input buffer is we now add a delay before we process user inputs.
	MaxInputBufferCommandFrames int = 15

	MaxStateBufferCommandFrames int = 5

	// The number of command frames on the server before a server update is sent to clients
	CommandFramesPerServerUpdate = 5

	// Animation
	AnimationMaxJointWeights = 4

	// Physics
	gravity float64 = 250
	// gravity float64 = 1

	SpatialPartitionNumPartitions int = 10
	SpatialPartitionDimensionSize int = 200

	// Debugging constants
	LatencyInjection = 0 * time.Millisecond

	// Mostly for debug rendering
	DefaultLineThickness float64 = 0.25
)
