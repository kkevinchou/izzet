package settings

type GameMode string

var (
	// dynamic settings loaded from config
	Width      int  = 0
	Height     int  = 0
	Fullscreen bool = false

	ShowImguiDemoWindow   = false
	RuntimeMaxTextureSize int
)

const (
	LoggingLevel       = 1
	Seed         int64 = 1234567

	PProfEnabled    bool = false
	PProfClientPort int  = 6060
	PProfServerPort int  = 6061

	// MSPerCommandFrame is the size of the simulation step for reading input,
	// physics, etc.
	MSPerCommandFrame int = 7
	FPS               int = 144

	// Animation
	MaxAnimationJointWeights = 4

	maxUInt32 uint32 = ^uint32(0)
	// we shift 8 bits since 8 bits are reserved for the alpha channel
	// the max id is used to indicate no entity was selected
	EmptyColorPickingID uint32 = maxUInt32 >> 8
)
