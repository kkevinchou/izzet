package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	_ "net/http/pprof"

	"github.com/kkevinchou/izzet/izzet"
	"github.com/kkevinchou/izzet/izzet/server"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/veandco/go-sdl2/sdl"
)

func init() {
	// We want to lock the main thread to this goroutine.  Otherwise,
	// SDL rendering will randomly panic
	//
	// For more details: https://github.com/golang/go/wiki/LockOSThread
	runtime.LockOSThread()
}

const (
	modeLocal  string = "LOCAL"
	modeClient string = "CLIENT"
	modeServer string = "SERVER"
)

type Game interface {
	Start()
}

func main() {
	configFile, err := os.Open("config.json")
	if err != nil {
		fmt.Printf("failed to load config.json, using defaults: %s\n", err)
	} else {
		configBytes, err := io.ReadAll(configFile)
		if err != nil {
			panic(err)
		}

		if err = configFile.Close(); err != nil {
			panic(err)
		}

		var configSettings Config
		err = json.Unmarshal(configBytes, &configSettings)
		if err != nil {
			panic(err)
		}

		loadConfig(configSettings)
	}

	if settings.Profile {
		go func() {
			http.ListenAndServe(":6868", nil)
		}()
	}

	client := false

	if len(os.Args) > 1 {
		mode := strings.ToUpper(os.Args[1])
		client = mode == "CLIENT"
		if mode != "SERVER" && mode != "CLIENT" {
			panic(fmt.Sprintf("unexpected mode %s", mode))
		}
	}

	if !client {
		go func() {
			serverApp := server.New("_assets", "shaders", "izzet_data.json")
			serverApp.Start()
		}()
	}

	app := izzet.New("_assets", "shaders", "izzet_data.json")
	app.Start()
	sdl.Quit()
}

func loadConfig(c Config) {
	settings.Width = c.Width
	settings.Height = c.Height
	settings.Fullscreen = c.Fullscreen
	settings.Profile = c.Profile
}

type Config struct {
	Width      int
	Height     int
	Fullscreen bool
	Profile    bool
}
