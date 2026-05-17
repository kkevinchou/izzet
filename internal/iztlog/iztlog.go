package iztlog

import (
	"log/slog"
)

var Logger *slog.Logger
var ClientLogger *slog.Logger
var ServerLogger *slog.Logger

func init() {
	Logger = slog.Default()
	ClientLogger = slog.Default()
	ServerLogger = slog.Default()
}

func SetLogger(logger *slog.Logger) {
	Logger = logger
}
func SetClientLogger(logger *slog.Logger) {
	ClientLogger = logger
}
func SetServerLogger(logger *slog.Logger) {
	ServerLogger = logger
}
