package pkg

import (
	"go.uber.org/zap"
)

// Logger is the interface that the loggers used by the library will use.
type Logger interface {
	Infof(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// ZapLogger is a logger using the zap logger underneath.
type ZapLogger struct {
	logger zap.SugaredLogger
}

// Infof logging with INFO level.
func (l *ZapLogger) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args)
}

// Debugf logging with DEBUG level.
func (l *ZapLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(format, args)
}

// Warningf logging with WARNING level.
func (l *ZapLogger) Warningf(format string, args ...interface{}) {
	l.logger.Warnf(format, args)
}

// Errorf logging with ERROR level.
func (l *ZapLogger) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args)
}

// NewLogger returns a new Logger using the zap logger underneath
func NewLogger(logger zap.SugaredLogger) *ZapLogger {
	return &ZapLogger{
		logger: logger,
	}
}

// NewDummyLogger returns a new Logger that doesn't log anything
func NewDummyLogger() *DummyLogger {
	return &DummyLogger{}
}

// DummyLogger is a logger that logs nothing
type DummyLogger struct {
	logger zap.SugaredLogger
}

// Infof logging with INFO level.
func (d *DummyLogger) Infof(format string, args ...interface{}) {}

// Debugf logging with DEBUG level.
func (d *DummyLogger) Debugf(format string, args ...interface{}) {}

// Warningf logging with WARNING level.
func (d *DummyLogger) Warningf(format string, args ...interface{}) {}

// Errorf logging with ERROR level.
func (d *DummyLogger) Errorf(format string, args ...interface{}) {}
