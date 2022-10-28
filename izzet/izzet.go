package izzet

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/kkevinchou/izzet/izzet/commandframe"
	"github.com/kkevinchou/izzet/izzet/directory"
	"github.com/kkevinchou/izzet/izzet/entitymanager"
	"github.com/kkevinchou/izzet/izzet/managers/eventbroker"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/spatialpartition"
	"github.com/kkevinchou/kitolib/input"
	"github.com/kkevinchou/kitolib/metrics"

	"github.com/kkevinchou/izzet/izzet/singleton"
	"github.com/kkevinchou/izzet/izzet/types"
)

type System interface {
	Name() string
	Update(delta time.Duration)
}

type RenderFunction func(delta time.Duration)

func emptyRenderFunction(delta time.Duration) {}

type Game struct {
	gameOver bool
	gameMode types.GameMode

	singleton        *singleton.Singleton
	entityManager    *entitymanager.EntityManager
	spatialPartition *spatialpartition.SpatialPartition
	systems          []System

	eventBroker     eventbroker.EventBroker
	metricsRegistry *metrics.MetricsRegistry

	inputPollingFn input.InputPoller

	// Client
	commandFrameHistory *commandframe.CommandFrameHistory
	focusedWindow       types.Window
	windowVisibility    map[types.Window]bool

	serverStats map[string]string
}

func NewBaseGame() *Game {
	g := &Game{
		gameMode:        types.GameModePlaying,
		singleton:       singleton.NewSingleton(),
		entityManager:   entitymanager.NewEntityManager(),
		eventBroker:     eventbroker.NewEventBroker(),
		metricsRegistry: metrics.New(),
		inputPollingFn:  input.NullInputPoller,
		focusedWindow:   types.WindowGame,
		windowVisibility: map[types.Window]bool{
			types.WindowGame: true,
		},
	}

	s := spatialpartition.NewSpatialPartition(g, settings.SpatialPartitionDimensionSize, settings.SpatialPartitionNumPartitions)
	g.spatialPartition = s
	return g
}

func (g *Game) Start() {
	var accumulator float64
	var renderAccumulator float64

	msPerFrame := float64(1000) / float64(settings.FPS)
	previousTimeStamp := float64(time.Now().UnixNano()) / 1000000

	frameCount := 0
	renderFunction := getRenderFunction()
	for !g.gameOver {
		now := float64(time.Now().UnixNano()) / 1000000
		delta := now - previousTimeStamp
		previousTimeStamp = now

		accumulator += delta
		renderAccumulator += delta

		runCount := 0
		timings := map[string]int{}
		for accumulator >= float64(settings.MSPerCommandFrame) {
			// input is handled once per command frame
			g.HandleInput(g.inputPollingFn())
			curTimings := g.runCommandFrame(time.Duration(settings.MSPerCommandFrame) * time.Millisecond)
			for k, v := range curTimings {
				timings[k] += v
			}

			// if timings["CollisionSystem"] != 0 {
			// 	fmt.Println(timings["CollisionSystem"])
			// }

			accumulator -= float64(settings.MSPerCommandFrame)
			runCount++
			// var timingsTotal int
			// for _, v := range timings {
			// 	timingsTotal += v
			// }
			// fmt.Println(timingsTotal)

		}

		// prevents lighting my CPU on fire
		if accumulator < float64(settings.MSPerCommandFrame)-10 {
			time.Sleep(5 * time.Millisecond)
		}

		if runCount > 1 {
			g.metricsRegistry.Inc("frameCatchup", 1)
		}

		if renderAccumulator >= msPerFrame {
			frameCount++
			g.metricsRegistry.Inc("fps", 1)
			start := time.Now()
			renderFunction(time.Duration(msPerFrame) * time.Millisecond)
			g.metricsRegistry.Inc("rendertime", float64(time.Since(start).Milliseconds()))
			renderAccumulator -= msPerFrame
		}
	}
}

func (g *Game) runCommandFrame(delta time.Duration) map[string]int {
	result := map[string]int{}
	g.singleton.CommandFrame++
	var total int
	for _, system := range g.systems {
		start := time.Now()
		system.Update(delta)
		systemTime := int(time.Since(start).Milliseconds())
		result[system.Name()] = systemTime
		total += systemTime
	}
	g.MetricsRegistry().Inc("frametime", float64(total))
	return result
}

func initSeed() {
	seed := settings.Seed
	fmt.Printf("initializing with seed %d ...\n", seed)
	rand.Seed(seed)
}

func getRenderFunction() RenderFunction {
	renderFunction := emptyRenderFunction
	d := directory.GetDirectory()
	renderSystem := d.RenderSystem()
	if renderSystem != nil {
		renderFunction = renderSystem.Render
	}

	return renderFunction
}
