package iztlog

import (
	"log/slog"
)

var Logger *slog.Logger

func init() {
	Logger = slog.Default()
}

func SetLogger(logger *slog.Logger) {
	Logger = logger
}
