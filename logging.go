package bcr

import (
	"log"

	"go.uber.org/zap"
)

// LogFunc is a function used for logging
type LogFunc func(template string, args ...interface{})

// Logger is a basic logger
type Logger struct {
	Debug LogFunc
	Info  LogFunc
	Error LogFunc
}

// NewNoopLogger returns a Logger that does nothing when its functions are called
func NewNoopLogger() *Logger {
	return &Logger{
		Debug: func(t string, args ...interface{}) {
			return
		},
		Info: func(t string, args ...interface{}) {
			return
		},
		Error: func(t string, args ...interface{}) {
			return
		},
	}
}

// wrapStdLog returns a function that calls log.Printf with the given prefix
func wrapStdLog(prefix string) LogFunc {
	return func(template string, args ...interface{}) {
		log.Printf(prefix+": "+template, args...)
	}
}

// NewStdlibLogger returns a Logger that wraps the standard library's "log" package
func NewStdlibLogger(debug bool) *Logger {
	if debug {
		return &Logger{
			Debug: wrapStdLog("DEBUG"),
			Info:  wrapStdLog("INFO"),
			Error: wrapStdLog("ERROR"),
		}
	}

	return &Logger{
		Debug: func(t string, args ...interface{}) {
			return
		},
		Info:  wrapStdLog("INFO"),
		Error: wrapStdLog("ERROR"),
	}
}

// NewZapLogger returns a Logger that wraps a SugaredLogger from go.uber.org/zap
func NewZapLogger(s *zap.SugaredLogger) *Logger {
	return &Logger{
		Debug: s.Debugf,
		Info:  s.Infof,
		Error: s.Errorf,
	}
}
