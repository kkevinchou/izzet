package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"net/http"
	_ "net/http/pprof"

	"github.com/kkevinchou/izzet/izzet/apputils"
	"github.com/kkevinchou/izzet/izzet/assets/assetslog"
	"github.com/kkevinchou/izzet/izzet/client"
	"github.com/kkevinchou/izzet/izzet/server"
	"github.com/kkevinchou/izzet/izzet/settings"
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
			http.ListenAndServe("localhost:6868", nil)
		}()
	}

	assetslog.SetLogger(log.EmptyLogger)

	mode := "CLIENT"

	if len(os.Args) > 1 {
		mode = strings.ToUpper(os.Args[1])
		if mode != "SERVER" && mode != "CLIENT" && mode != "HEADLESS" {
			panic(fmt.Sprintf("unexpected mode %s", mode))
		}
	}

	defaultProject := settings.DefaultProject

	if mode == "SERVER" {
		clientApp := client.New("shaders", config, defaultProject)
		if err != nil {
			panic(err)
		}
		clientApp.Start()
	} else if mode == "HEADLESS" {
		started := make(chan bool)
		go func() {
			<-started
		}()
		serverApp := server.NewWithFile(apputils.PathToProjectFile(defaultProject), settings.DefaultProject)
		serverApp.Start(started, make(chan bool))
	} else if mode == "CLIENT" {
		config.Fullscreen = false
		config.Profile = false
		clientApp := client.New("shaders", config, "")
		clientApp.Connect()
		clientApp.Start()
	}

	sdl.Quit()
}
