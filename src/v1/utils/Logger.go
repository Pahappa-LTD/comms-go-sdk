package utils

import "sync"

// Logger is what the SDK writes its log messages to. Each method takes a message followed by
// alternating key/value pairs, so *slog.Logger (Go 1.21+) and hclog.Logger can be used as they are;
// other logging libraries need a small adapter.
type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
}

type nopLogger struct{}

func (nopLogger) Debug(string, ...interface{}) {}
func (nopLogger) Info(string, ...interface{})  {}
func (nopLogger) Warn(string, ...interface{})  {}
func (nopLogger) Error(string, ...interface{}) {}

var (
	loggerMu sync.RWMutex
	logger   Logger = nopLogger{}
)

// SetLogger sets the logger the SDK writes to. Passing nil makes the SDK silent again (the default).
func SetLogger(l Logger) {
	if l == nil {
		l = nopLogger{}
	}
	loggerMu.Lock()
	logger = l
	loggerMu.Unlock()
}

// Log returns the logger the SDK writes to.
func Log() Logger {
	loggerMu.RLock()
	defer loggerMu.RUnlock()
	return logger
}
