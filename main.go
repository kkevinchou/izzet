package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strconv"
	"strings"

	"net/http"
	_ "net/http/pprof"

	"github.com/kkevinchou/izzet/internal/iztlog"
	"github.com/kkevinchou/izzet/izzet/client"
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

	logHandlerOptions := &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			if attr.Key == slog.TimeKey && attr.Value.Kind() == slog.KindTime {
				attr.Value = slog.StringValue(attr.Value.Time().Local().Format("15:04:05.000"))
			}
			return attr
		},
	}

	closeLogs := configureLoggers(logsEnabled, logHandlerOptions)
	defer closeLogs()

	iztlog.Logger.Info("====================================================================================")
	iztlog.Logger.Info("IZZET SESSION START")
	iztlog.Logger.Info("====================================================================================")

	iztlog.ClientLogger.Info("====================================================================================")
	iztlog.ClientLogger.Info("IZZET CLIENT SESSION START")
	iztlog.ClientLogger.Info("====================================================================================")

	iztlog.ServerLogger.Info("====================================================================================")
	iztlog.ServerLogger.Info("IZZET SERVER SESSION START")
	iztlog.ServerLogger.Info("====================================================================================")

	if mode == "SERVER" {
		clientApp := client.New("shaders", config)
		if err != nil {
			panic(err)
		}
		clientApp.Start()
	} else if mode == "HEADLESS" {
	} else if mode == "CLIENT" {
		clientApp := client.New("shaders", config)
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

func configureLoggers(logsEnabled bool, logHandlerOptions *slog.HandlerOptions) func() {
	if !logsEnabled {
		f, err := os.OpenFile("_logs_garbage.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}

		logger := slog.New(slog.NewJSONHandler(f, logHandlerOptions))
		iztlog.SetLogger(logger)
		iztlog.SetClientLogger(logger)
		iztlog.SetServerLogger(logger)

		return func() {
			if err := f.Close(); err != nil {
				panic(err)
			}
		}
	}

	appLog, err := os.OpenFile("app.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	iztlog.SetLogger(slog.New(slog.NewJSONHandler(appLog, logHandlerOptions)))

	clientLog, err := os.OpenFile("_logs_client.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	iztlog.SetClientLogger(slog.New(slog.NewJSONHandler(clientLog, logHandlerOptions)))

	serverLog, err := os.OpenFile("_logs_server.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	iztlog.SetServerLogger(slog.New(slog.NewJSONHandler(serverLog, logHandlerOptions)))

	return func() {
		if err := appLog.Close(); err != nil {
			panic(err)
		}
		if err := clientLog.Close(); err != nil {
			panic(err)
		}
		if err := serverLog.Close(); err != nil {
			panic(err)
		}
	}
}
