package izzet

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/kkevinchou/izzet/izzet/edithistory"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/modellibrary"
	"github.com/kkevinchou/izzet/izzet/observers"
	"github.com/kkevinchou/izzet/izzet/serialization"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/systems"
	"github.com/kkevinchou/izzet/izzet/world"
	"github.com/kkevinchou/kitolib/assets"
	"github.com/kkevinchou/kitolib/metrics"
)

func NewServer(assetsDirectory, shaderDirectory, dataFilePath string) *Izzet {
	initSeed()
	g := &Izzet{isServer: true, playerIDGenerator: 100000}
	g.initSettings()

	g.assetManager = assets.NewAssetManager(assetsDirectory, false)
	g.modelLibrary = modellibrary.New(false)
	// g.appMode = app.AppModeEditor

	// g.camera = &camera.Camera{
	// 	Position:    mgl64.Vec3{-82, 230, 95},
	// 	Orientation: mgl64.QuatIdent(),
	// }

	start := time.Now()

	g.world = world.New(map[int]*entities.Entity{})

	fmt.Println(time.Since(start), "spatial partition done")

	g.entities = map[int]*entities.Entity{}
	// data := izzetdata.LoadData(dataFilePath)
	// g.setupAssets(g.assetManager, g.modelLibrary, data)
	g.serializer = serialization.New(g, g.world)
	g.metricsRegistry = metrics.New()
	g.physicsObserver = observers.NewPhysicsObserver()

	// THINGS TO DELETE AFTER DEBUGGING
	g.editHistory = edithistory.New()

	g.serverModeSystems = append(g.serverModeSystems, &systems.MovementSystem{})
	g.serverModeSystems = append(g.serverModeSystems, &systems.PhysicsSystem{Observer: g.physicsObserver})

	// g.setupEntities(data)
	g.LoadWorld("cubes")

	fmt.Println(time.Since(start), "to start up systems")

	return g
}

func (g *Izzet) StartServer() {
	host := "0.0.0.0"
	port := "7878"
	listener, err := net.Listen("tcp", host+":"+port)
	if err != nil {
		panic(err)
	}
	_ = listener

	fmt.Println("listening on " + host + ":" + port)

	go func() {
		defer listener.Close()
		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println("error accepting a connection on the listener:", err.Error())
				continue
			}

			g.playerIDLock.Lock()
			id := g.playerIDGenerator
			g.playerIDGenerator += 1
			g.playerIDLock.Unlock()

			// message, err := s.createAcceptMessage(id)
			// if err != nil {
			// 	fmt.Println(err)
			// 	continue
			// }
			// sendMessage(conn, message)
			// if err != nil {
			// 	fmt.Println("error sending accept message:", err.Error())
			// 	continue
			// }

			encoder := json.NewEncoder(conn)
			err = encoder.Encode(id)
			if err != nil {
				fmt.Println("error with incoming message: %w", err)
			}

			// select {
			// case s.incomingConnections <- &Connection{ID: id, Connection: conn}:
			// default:
			// 	panic("incomingConnections queue full")
			// }
		}

	}()

	var accumulator float64

	// msPerFrame := float64(1000) / float64(60)
	previousTimeStamp := float64(time.Now().UnixNano()) / 1000000

	commandFrameCountBeforeRender := 0
	for !g.gameOver {
		now := float64(time.Now().UnixNano()) / 1000000
		delta := now - previousTimeStamp
		previousTimeStamp = now

		accumulator += delta

		currentLoopCommandFrames := 0
		for accumulator >= float64(settings.MSPerCommandFrame) {
			start := time.Now()
			g.runCommandFrameServer(time.Duration(settings.MSPerCommandFrame) * time.Millisecond)
			commandFrameNanos := time.Since(start).Nanoseconds()
			g.MetricsRegistry().Inc("command_frame_nanoseconds", float64(commandFrameNanos))
			g.world.IncrementCommandFrameCount()
			commandFrameCountBeforeRender += 1

			accumulator -= float64(settings.MSPerCommandFrame)
			currentLoopCommandFrames++
			if currentLoopCommandFrames > settings.MaxCommandFramesPerLoop {
				accumulator = 0
			}
		}
	}
}
