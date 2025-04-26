package common

import (
	"fmt"

	logging "github.com/ipfs/go-log/v2"
)

type Logger interface {
	Debugw(msg string, keysAndValues ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Warnw(msg string, err error, keysAndValues ...interface{})
	Errorw(msg string, err error, keysAndValues ...interface{})

	// Add these methods to match the usage in the codebase
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})

	// Add Debug method to satisfy StandardLogger interface
	Debug(args ...interface{})
}

func printMessage(level, msg string, keysAndValues ...interface{}) {
	msg = fmt.Sprintf(msg, keysAndValues...)
	fmt.Printf("%s: %s\n", level, msg)
}

func printMessageWithError(level, msg string, err error, keysAndValues ...interface{}) {
	msg = fmt.Sprintf(msg, keysAndValues...)
	fmt.Printf("%s: %s (err = %s)\n", level, msg, err.Error())
}

type ConsoleLogger struct{}

func (c ConsoleLogger) Debugw(msg string, keysAndValues ...interface{}) {
	printMessage("DEBUG", msg, keysAndValues...)
}

func (c ConsoleLogger) Infow(msg string, keysAndValues ...interface{}) {
	printMessage("INFO", msg, keysAndValues...)
}

func (c ConsoleLogger) Warnw(msg string, err error, keysAndValues ...interface{}) {
	printMessageWithError("WARN", msg, err, keysAndValues...)
}

func (c ConsoleLogger) Errorw(msg string, err error, keysAndValues ...interface{}) {
	printMessageWithError("ERROR", msg, err, keysAndValues...)
}

// Implement the additional methods
func (c ConsoleLogger) Debugf(format string, args ...interface{}) {
	printMessage("DEBUG", format, args...)
}

func (c ConsoleLogger) Infof(format string, args ...interface{}) {
	printMessage("INFO", format, args...)
}

func (c ConsoleLogger) Warnf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("WARN: %s\n", msg)
}

func (c ConsoleLogger) Errorf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("ERROR: %s\n", msg)
}

func (c ConsoleLogger) Debug(args ...interface{}) {
	fmt.Printf("DEBUG: %s\n", fmt.Sprint(args...))
}

// ZapLoggerAdapter adapts a ZapEventLogger to our Logger interface
type ZapLoggerAdapter struct {
	logger *logging.ZapEventLogger
}

func NewZapLoggerAdapter(logger *logging.ZapEventLogger) Logger {
	return &ZapLoggerAdapter{logger: logger}
}

// Forward all methods to the underlying ZapEventLogger
func (z *ZapLoggerAdapter) Debugw(msg string, keysAndValues ...interface{}) {
	z.logger.Debugw(msg, keysAndValues...)
}

func (z *ZapLoggerAdapter) Infow(msg string, keysAndValues ...interface{}) {
	z.logger.Infow(msg, keysAndValues...)
}

func (z *ZapLoggerAdapter) Warnw(msg string, err error, keysAndValues ...interface{}) {
	keysAndValues = append(keysAndValues, "error", err)
	z.logger.Warnw(msg, keysAndValues...)
}

func (z *ZapLoggerAdapter) Errorw(msg string, err error, keysAndValues ...interface{}) {
	keysAndValues = append(keysAndValues, "error", err)
	z.logger.Errorw(msg, keysAndValues...)
}

func (z *ZapLoggerAdapter) Debugf(format string, args ...interface{}) {
	z.logger.Debugf(format, args...)
}

func (z *ZapLoggerAdapter) Infof(format string, args ...interface{}) {
	z.logger.Infof(format, args...)
}

func (z *ZapLoggerAdapter) Warnf(format string, args ...interface{}) {
	z.logger.Warnf(format, args...)
}

func (z *ZapLoggerAdapter) Errorf(format string, args ...interface{}) {
	z.logger.Errorf(format, args...)
}

func (z *ZapLoggerAdapter) Debug(args ...interface{}) {
	z.logger.Debug(args...)
}

// StandardLoggerAdapter adapts our Logger to implement logging.StandardLogger interface
type StandardLoggerAdapter struct {
	Logger
}

// These methods implement the full logging.StandardLogger interface
func (s StandardLoggerAdapter) Warn(args ...interface{}) {
	s.Warnf("%s", fmt.Sprint(args...))
}

func (s StandardLoggerAdapter) Info(args ...interface{}) {
	s.Infof("%s", fmt.Sprint(args...))
}

func (s StandardLoggerAdapter) Error(args ...interface{}) {
	s.Errorf("%s", fmt.Sprint(args...))
}

func (s StandardLoggerAdapter) Fatal(args ...interface{}) {
	s.Errorf("FATAL: %s", fmt.Sprint(args...))
}

func (s StandardLoggerAdapter) Fatalf(format string, args ...interface{}) {
	s.Errorf("FATAL: "+format, args...)
}

func (s StandardLoggerAdapter) Panic(args ...interface{}) {
	s.Errorf("PANIC: %s", fmt.Sprint(args...))
}

func (s StandardLoggerAdapter) Panicf(format string, args ...interface{}) {
	s.Errorf("PANIC: "+format, args...)
}

func (s StandardLoggerAdapter) Debugw(msg string, keysAndValues ...interface{}) {
	s.Logger.Debugw(msg, keysAndValues...)
}

func (s StandardLoggerAdapter) Infow(msg string, keysAndValues ...interface{}) {
	s.Logger.Infow(msg, keysAndValues...)
}

func (s StandardLoggerAdapter) Warnw(msg string, keysAndValues ...interface{}) {
	args := make([]interface{}, 0, len(keysAndValues)+2)
	args = append(args, "msg", msg)
	args = append(args, keysAndValues...)
	s.Logger.Warnw(msg, nil, args...)
}

func (s StandardLoggerAdapter) Errorw(msg string, keysAndValues ...interface{}) {
	args := make([]interface{}, 0, len(keysAndValues)+2)
	args = append(args, "msg", msg)
	args = append(args, keysAndValues...)
	s.Logger.Errorw(msg, nil, args...)
}

func (s StandardLoggerAdapter) Fatalw(msg string, keysAndValues ...interface{}) {
	args := make([]interface{}, 0, len(keysAndValues)+2)
	args = append(args, "msg", msg)
	args = append(args, keysAndValues...)
	s.Logger.Errorw("FATAL: "+msg, nil, args...)
}

func (s StandardLoggerAdapter) Panicw(msg string, keysAndValues ...interface{}) {
	args := make([]interface{}, 0, len(keysAndValues)+2)
	args = append(args, "msg", msg)
	args = append(args, keysAndValues...)
	s.Logger.Errorw("PANIC: "+msg, nil, args...)
}

// Create a new StandardLoggerAdapter from our Logger
func NewStandardLoggerAdapter(logger Logger) logging.StandardLogger {
	return &StandardLoggerAdapter{Logger: logger}
}
