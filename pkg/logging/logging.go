package logging

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"sync"
	"time"
)

var once sync.Once

type Loggers struct {
	module string
	method string
}

// NewLoggers returns new loggers with module anf method
func NewLoggers(module, method string) *Loggers {
	once.Do(func() {
		log.Logger = log.Output(zerolog.SyncWriter(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}))
	})

	return &Loggers{module: module, method: method}
}

// addMetadata add module and method
func (l *Loggers) addMetadata(e *zerolog.Event) *zerolog.Event {
	return e.Str("module", l.module).Str("method", l.method)
}

// DebugLog returns log.Debug()
func (l *Loggers) DebugLog() *zerolog.Event {
	return l.addMetadata(log.Debug())
}

// InfoLog returns log.Info()
func (l *Loggers) InfoLog() *zerolog.Event {
	return l.addMetadata(log.Info())
}

// WarnLog returns log.Warn()
func (l *Loggers) WarnLog() *zerolog.Event {
	return l.addMetadata(log.Warn())
}

// ErrorLog returns log.Error()
func (l *Loggers) ErrorLog() *zerolog.Event {
	return l.addMetadata(log.Error())
}

// FatalLog returns log.Fatal()
func (l *Loggers) FatalLog() *zerolog.Event {
	return l.addMetadata(log.Fatal())
}
