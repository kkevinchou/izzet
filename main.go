package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"

	"net/http"
	_ "net/http/pprof"

	"github.com/kkevinchou/izzet/izzet/client"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/veandco/go-sdl2/sdl"
)

const logDir = "logs"

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
	mode, logsEnabled := parseCommandLine(os.Args[1:])

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

	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic(err)
	}

	if mode == "SERVER" {
		config.Fullscreen = false
		clientApp := client.New("shaders", config, logsEnabled)
		if err != nil {
			panic(err)
		}
		clientApp.Start()
	} else if mode == "HEADLESS" {
	} else if mode == "CLIENT" {
		clientApp := client.New("shaders", config, logsEnabled)
		clientApp.Connect()
		clientApp.Start()
	}

	sdl.Quit()
}

func parseCommandLine(args []string) (string, bool) {
	mode := modeClient
	logsEnabled := true

	for _, arg := range args {
		if strings.HasPrefix(arg, "--logs=") {
			value, err := strconv.ParseBool(strings.TrimPrefix(arg, "--logs="))
			if err != nil {
				panic(fmt.Sprintf("unexpected --logs value %q", strings.TrimPrefix(arg, "--logs=")))
			}
			logsEnabled = value
			continue
		}

		if strings.HasPrefix(arg, "--") {
			panic(fmt.Sprintf("unexpected flag %s", arg))
		}

		mode = strings.ToUpper(arg)
		if mode != modeServer && mode != modeClient && mode != "HEADLESS" {
			panic(fmt.Sprintf("unexpected mode %s", mode))
		}
	}

	return mode, logsEnabled
}
