package clientsystem

import (
	"log/slog"
	"net"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/input"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/render/context"
	"github.com/kkevinchou/izzet/izzet/runtimeconfig"
	"github.com/kkevinchou/izzet/izzet/serverstats"
	"github.com/kkevinchou/izzet/izzet/world"
)

type App interface {
	AssetManager() *assets.AssetManager
	NetworkMessagesChannel() chan network.MessageTransport
	GetPlayerID() int
	CommandFrame() int
	IsConnected() bool
	IsClient() bool
	IsServer() bool
	Logger() *slog.Logger
	GetPlayerConnection() net.Conn
	GetPlayerEntity() *entity.Entity
	GetPlayerCamera() *entity.Entity
	GetCommandFrameHistory() *CommandFrameHistory
	Client() network.IzzetClient
	StateBuffer() *StateBuffer
	GetFrameInput() input.Input
	GetFrameInputPtr() *input.Input
	SetServerStats(stats serverstats.ServerStats)
	RuntimeConfig() *runtimeconfig.RuntimeConfig
	World() *world.GameWorld
	PredictionDebugLogging() bool
	SetPredictionDebugLogging(value bool)
	MouseCaptured() bool
	SetMouseCaptured(capture bool)
	SetCapturedMouseOrigin(x, y int32)
	SceneSize() (int, int)
	CameraViewerContext() context.ViewerContext
	IntersectRayWithEntities(position, dir mgl64.Vec3) (mgl64.Vec3, bool)
}
