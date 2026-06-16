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

func SetClientLogger(logger *slog.Logger, commandFrameFn func() int) {
	ClientLogger = withCommandFrame(logger, "cf", commandFrameFn)
}
func SetServerLogger(logger *slog.Logger, commandFrameFn func() int) {
	ServerLogger = withCommandFrame(logger, "gcf", commandFrameFn)
}

func withCommandFrame(logger *slog.Logger, key string, fn func() int) *slog.Logger {
	return slog.New(IztLogHandler{
		handler: logger.Handler(),
		key:     key,
		fn:      fn,
	})
}

type IztLogHandler struct {
	handler slog.Handler
	key     string
	fn      func() int
}

func (h IztLogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h IztLogHandler) Handle(ctx context.Context, record slog.Record) error {
	record = withRecordAttr(record, slog.Int(h.key, h.fn()))
	return h.handler.Handle(ctx, record)
}

func (h IztLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return IztLogHandler{handler: h.handler.WithAttrs(attrs)}
}

func (h IztLogHandler) WithGroup(name string) slog.Handler {
	return IztLogHandler{handler: h.handler.WithGroup(name)}
}

func withRecordAttr(record slog.Record, nextAttr slog.Attr) slog.Record {
	next := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	record.Attrs(func(attr slog.Attr) bool {
		if attr.Key != nextAttr.Key {
			next.AddAttrs(attr)
		}
		return true
	})
	next.AddAttrs(nextAttr)
	return next
}
