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

	"github.com/kkevinchou/izzet/izzet/client"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/kitolib/assets/assetslog"
	"github.com/kkevinchou/kitolib/log"
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
	config := settings.NewConfig()

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

		err = json.Unmarshal(configBytes, &config)
		if err != nil {
			panic(err)
		}
	}

	if config.Profile {
		go func() {
			http.ListenAndServe(":6868", nil)
		}()
	}

	assetslog.SetLogger(log.EmptyLogger)

	isServer := false

	if len(os.Args) > 1 {
		mode := strings.ToUpper(os.Args[1])
		isServer = mode == "SERVER"
		if mode != "SERVER" && mode != "CLIENT" {
			panic(fmt.Sprintf("unexpected mode %s", mode))
		}
	}

	if isServer {
		clientApp := client.New("_assets", "shaders", "izzet_data.json", config)
		clientApp.StartAsyncServer()
		err := clientApp.Connect()
		if err != nil {
			fmt.Println(err)
		}
		clientApp.Start()
	} else {
		config.Width = 854
		config.Height = 480
		config.Fullscreen = false
		config.Profile = false
		clientApp := client.New("_assets", "shaders", "izzet_data.json", config)
		clientApp.Connect()
		clientApp.Start()
	}

	sdl.Quit()
}
