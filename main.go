package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"

	_ "net/http/pprof"

	"github.com/kkevinchou/izzet/izzet"
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
	// memory ballast
	ballast := make([]byte, 1<<34)
	_ = ballast

	configFile, err := os.Open("config.json")
	if err != nil {
		fmt.Printf("failed to load config.json, using defaults: %s\n", err)
	} else {
		configBytes, err := io.ReadAll(configFile)
		if err != nil {
			fmt.Println(err)
			return
		}

		if err = configFile.Close(); err != nil {
			fmt.Println(err)
			return
		}

		var configSettings Config
		err = json.Unmarshal(configBytes, &configSettings)
		if err != nil {
			fmt.Println(err)
			return
		}

		loadConfig(configSettings)
	}

	var mode string = modeClient
	if len(os.Args) > 1 {
		mode = strings.ToUpper(os.Args[1])
		if mode != modeLocal && mode != modeClient && mode != modeServer {
			panic(fmt.Sprintf("unexpected mode %s", mode))
		}
	}

	if settings.PProfEnabled {
		go func() {
			if mode == modeClient {
				log.Println(http.ListenAndServe(fmt.Sprintf("localhost:%d", settings.PProfClientPort), nil))
			} else {
				log.Println(http.ListenAndServe(fmt.Sprintf("localhost:%d", settings.PProfServerPort), nil))
			}
		}()
	}

	app := izzet.New("_assets", "shaders")
	app.Start()
	sdl.Quit()
}

func loadConfig(c Config) {
	settings.Width = c.Width
	settings.Height = c.Height
	settings.Fullscreen = c.Fullscreen
}

type Config struct {
	ServerIP   string
	ServerPort int
	Mode       string
	Width      int
	Height     int
	Fullscreen bool
}
