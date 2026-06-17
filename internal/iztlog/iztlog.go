package iztlog

import (
	"context"
	"log/slog"
)

var ClientLogger *slog.Logger
var ServerLogger *slog.Logger

func init() {
	ClientLogger = slog.Default()
	ServerLogger = slog.Default()
}

func SetClientLogger(logger *slog.Logger, commandFrameFn func() int, playerIDFn func() int) {
	ClientLogger = withClientAttrs(logger, commandFrameFn, playerIDFn)
}
func SetServerLogger(logger *slog.Logger, commandFrameFn func() int) {
	ServerLogger = withServerAttrs(logger, commandFrameFn)
}

func withClientAttrs(logger *slog.Logger, commandFrameFn func() int, playerIDFn func() int) *slog.Logger {
	return slog.New(IztLogHandler{
		handler:         logger.Handler(),
		commandFrameKey: "cf",
		commandFrameFn:  commandFrameFn,
		playerIDFn:      playerIDFn,
	})
}

func withServerAttrs(logger *slog.Logger, commandFrameFn func() int) *slog.Logger {
	return slog.New(IztLogHandler{
		handler:         logger.Handler(),
		commandFrameKey: "gcf",
		commandFrameFn:  commandFrameFn,
	})
}

type IztLogHandler struct {
	handler         slog.Handler
	commandFrameKey string
	commandFrameFn  func() int
	playerIDFn      func() int
}

func (h IztLogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h IztLogHandler) Handle(ctx context.Context, record slog.Record) error {
	var attrs []slog.Attr
	if h.commandFrameFn != nil {
		attrs = append(attrs, slog.Int(h.commandFrameKey, h.commandFrameFn()))
	}
	if h.playerIDFn != nil {
		attrs = append(attrs, slog.Int("player id", h.playerIDFn()))
	}

	record = withRecordAttrs(record, attrs...)
	return h.handler.Handle(ctx, record)
}

func (h IztLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return IztLogHandler{
		handler:         h.handler.WithAttrs(attrs),
		commandFrameKey: h.commandFrameKey,
		commandFrameFn:  h.commandFrameFn,
		playerIDFn:      h.playerIDFn,
	}
}

func (h IztLogHandler) WithGroup(name string) slog.Handler {
	return IztLogHandler{
		handler:         h.handler.WithGroup(name),
		commandFrameKey: h.commandFrameKey,
		commandFrameFn:  h.commandFrameFn,
		playerIDFn:      h.playerIDFn,
	}
}

func withRecordAttrs(record slog.Record, nextAttrs ...slog.Attr) slog.Record {
	next := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	record.Attrs(func(attr slog.Attr) bool {
		if !hasAttrKey(nextAttrs, attr.Key) {
			next.AddAttrs(attr)
		}
		return true
	})
	next.AddAttrs(nextAttrs...)
	return next
}

func hasAttrKey(attrs []slog.Attr, key string) bool {
	for _, attr := range attrs {
		if attr.Key == key {
			return true
		}
	}
	return false
}
