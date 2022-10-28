package main

import (
	"fmt"
	"os"

	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/incubator/imguiexample/example"
	"github.com/kkevinchou/izzet/incubator/imguiexample/platforms"
	"github.com/kkevinchou/izzet/incubator/imguiexample/renderers"
)

func main() {
	context := imgui.CreateContext(nil)
	defer context.Destroy()
	io := imgui.CurrentIO()

	platform, err := platforms.NewSDL(io, platforms.SDLClientAPIOpenGL4)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}
	defer platform.Dispose()

	renderer, err := renderers.NewOpenGL4(io)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}
	defer renderer.Dispose()

	example.Run(platform, renderer)
}
