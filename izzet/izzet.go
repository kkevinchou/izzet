package izzet

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/kitolib/input"
	"github.com/kkevinchou/kitolib/metrics"
	"github.com/kkevinchou/kitolib/shaders"
	"github.com/veandco/go-sdl2/sdl"
)

type Izzet struct {
	gameOver        bool
	metricsRegistry *metrics.MetricsRegistry
	platform        *input.SDLPlatform
	window          *sdl.Window

	shaderManager *shaders.ShaderManager
}

func New(assetsDirectory, shaderDirectory string) *Izzet {
	g := &Izzet{}
	initSeed()

	window, err := initializeOpenGL(settings.Width, settings.Height, settings.Fullscreen)
	if err != nil {
		panic(err)
	}
	imgui.CreateContext(nil)
	imguiIO := imgui.CurrentIO()
	g.platform = input.NewSDLPlatform(window, imguiIO)
	g.window = window
	g.shaderManager = shaders.NewShaderManager(shaderDirectory)

	var data int32
	gl.GetIntegerv(gl.MAX_TEXTURE_SIZE, &data)
	settings.RuntimeMaxTextureSize = int(data)

	// compileShaders(g.shaderManager)

	return g
}

func (g *Izzet) Start() {
	var accumulator float64
	var renderAccumulator float64

	msPerFrame := float64(1000) / float64(settings.FPS)
	previousTimeStamp := float64(time.Now().UnixNano()) / 1000000

	frameCount := 0
	for !g.gameOver {
		now := float64(time.Now().UnixNano()) / 1000000
		delta := now - previousTimeStamp
		previousTimeStamp = now

		accumulator += delta
		renderAccumulator += delta

		runCount := 0
		for accumulator >= float64(settings.MSPerCommandFrame) {
			g.HandleInput(g.platform.PollInput())
			g.runCommandFrame(time.Duration(settings.MSPerCommandFrame) * time.Millisecond)

			accumulator -= float64(settings.MSPerCommandFrame)
			runCount++
		}

		// prevents lighting my CPU on fire
		if accumulator < float64(settings.MSPerCommandFrame)-10 {
			time.Sleep(5 * time.Millisecond)
		}

		if renderAccumulator >= msPerFrame {
			frameCount++
			// renderFunction(time.Duration(msPerFrame) * time.Millisecond)
			initOpenGLRenderSettings()
			g.window.GLSwap()
			renderAccumulator -= msPerFrame
		}
	}
}

func (g *Izzet) runCommandFrame(delta time.Duration) map[string]int {
	result := map[string]int{}
	// var total int
	// for _, system := range g.systems {
	// 	start := time.Now()
	// 	system.Update(delta)
	// 	systemTime := int(time.Since(start).Milliseconds())
	// 	result[system.Name()] = systemTime
	// 	total += systemTime
	// }
	return result
}

func initSeed() {
	seed := settings.Seed
	fmt.Printf("initializing with seed %d ...\n", seed)
	rand.Seed(seed)
}
