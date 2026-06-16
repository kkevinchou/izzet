package iztlog

import (
	"log/slog"
)

var ClientLogger *slog.Logger
var ServerLogger *slog.Logger

func init() {
	ClientLogger = slog.Default()
	ServerLogger = slog.Default()
}

func SetClientLogger(logger *slog.Logger) {
	ClientLogger = logger
}
func SetServerLogger(logger *slog.Logger) {
	ServerLogger = logger
}
