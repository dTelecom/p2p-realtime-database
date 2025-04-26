package logger

import (
	"fmt"

	"github.com/dTelecom/p2p-realtime-database/internal/common"
)

// SimpleLogger represents a minimal logger interface that most logger implementations provide
type SimpleLogger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
}

// LoggerAdapter adapts a simpler logger to work with our internal package
type LoggerAdapter struct {
	logger SimpleLogger
}

// NewLoggerAdapter creates a new adapter that wraps a simpler logger
// and makes it compatible with our internal Logger interface
func NewLoggerAdapter(logger SimpleLogger) common.Logger {
	return &LoggerAdapter{logger: logger}
}

// Implement all required methods for common.Logger interface
func (a *LoggerAdapter) Debugw(msg string, keysAndValues ...interface{}) {
	if len(keysAndValues) > 0 {
		a.logger.Debug(fmt.Sprintf("%s %v", msg, keysAndValues))
	} else {
		a.logger.Debug(msg)
	}
}

func (a *LoggerAdapter) Infow(msg string, keysAndValues ...interface{}) {
	if len(keysAndValues) > 0 {
		a.logger.Info(fmt.Sprintf("%s %v", msg, keysAndValues))
	} else {
		a.logger.Info(msg)
	}
}

func (a *LoggerAdapter) Warnw(msg string, err error, keysAndValues ...interface{}) {
	var errStr string
	if err != nil {
		errStr = fmt.Sprintf(" (err=%s)", err.Error())
	}

	if len(keysAndValues) > 0 {
		a.logger.Warn(fmt.Sprintf("%s%s %v", msg, errStr, keysAndValues))
	} else {
		a.logger.Warn(fmt.Sprintf("%s%s", msg, errStr))
	}
}

func (a *LoggerAdapter) Errorw(msg string, err error, keysAndValues ...interface{}) {
	var errStr string
	if err != nil {
		errStr = fmt.Sprintf(" (err=%s)", err.Error())
	}

	if len(keysAndValues) > 0 {
		a.logger.Error(fmt.Sprintf("%s%s %v", msg, errStr, keysAndValues))
	} else {
		a.logger.Error(fmt.Sprintf("%s%s", msg, errStr))
	}
}

func (a *LoggerAdapter) Debugf(format string, args ...interface{}) {
	a.logger.Debug(fmt.Sprintf(format, args...))
}

func (a *LoggerAdapter) Infof(format string, args ...interface{}) {
	a.logger.Info(fmt.Sprintf(format, args...))
}

func (a *LoggerAdapter) Warnf(format string, args ...interface{}) {
	a.logger.Warn(fmt.Sprintf(format, args...))
}

func (a *LoggerAdapter) Errorf(format string, args ...interface{}) {
	a.logger.Error(fmt.Sprintf(format, args...))
}

func (a *LoggerAdapter) Debug(args ...interface{}) {
	a.logger.Debug(args...)
}
