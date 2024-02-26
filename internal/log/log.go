package log

import "log"

type KitoLogger interface {
	Println(v ...any)
	Printf(format string, v ...any)
}

var Logger KitoLogger

type StdOutLogger struct {
	internalLogger *log.Logger
}

func NewStdOutLogger() *StdOutLogger {
	return &StdOutLogger{
		internalLogger: log.Default(),
	}
}

func (l *StdOutLogger) Println(v ...any) {
	l.internalLogger.Println(v...)
}

func (l *StdOutLogger) Printf(format string, v ...any) {
	l.internalLogger.Printf(format, v...)
}

var EmptyLogger KitoLogger

type emptyLogger struct {
}

func (l *emptyLogger) Println(v ...any) {
}

func (l *emptyLogger) Printf(format string, v ...any) {
}

func init() {
	EmptyLogger = &emptyLogger{}
}
