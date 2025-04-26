package p2p_database

import (
	"github.com/dTelecom/p2p-realtime-database/internal/common"
	"github.com/dTelecom/p2p-realtime-database/logger"
)

// NewLoggerAdapter creates a new adapter that can convert a simple logger to the Logger interface required by Connect.
// The simple logger only needs to implement Debug, Info, Warn, and Error methods.
func NewLoggerAdapter(simpleLogger logger.SimpleLogger) common.Logger {
	return logger.NewLoggerAdapter(simpleLogger)
}

// NewLivekitLoggerAdapter creates a new adapter specifically for LiveKit loggers.
// This adapter works with loggers that have the Debugw, Infow, Warnw, and Errorw methods.
func NewLivekitLoggerAdapter(livekitLogger logger.LivekitLogger) common.Logger {
	return logger.NewLivekitLoggerAdapter(livekitLogger)
}

// NewConsoleLogger creates a new simple console logger that can be used with Connect.
func NewConsoleLogger() common.Logger {
	return new(common.ConsoleLogger)
}
