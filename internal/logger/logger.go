package logger

import (
	"fmt"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ParseLevel converts a string log level to a zapcore.Level.
// Valid levels: debug, info, warn, error, dpanic, panic, fatal
// Defaults to InfoLevel if the string is invalid.
func ParseLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "dpanic":
		return zapcore.DPanicLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// New creates a new zap logger configured for JSON output to stdout
// with a common time format (RFC3339) and the specified log level.
func New(level zapcore.Level) (*zap.Logger, error) {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	config.EncodeLevel = zapcore.LowercaseLevelEncoder

	encoder := zapcore.NewJSONEncoder(config)
	core := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)

	return zap.New(core), nil
}

// MustNew creates a new zap logger with the specified log level and panics if initialization fails.
// This is useful for initialization at program startup.
func MustNew(level zapcore.Level) *zap.Logger {
	logger, err := New(level)
	if err != nil {
		panic(fmt.Errorf("failed to create logger: %w", err))
	}
	return logger
}

// MustNewFromString creates a new zap logger from a string log level and panics if initialization fails.
// Valid levels: debug, info, warn, error, dpanic, panic, fatal
// Defaults to InfoLevel if the string is invalid.
func MustNewFromString(level string) *zap.Logger {
	return MustNew(ParseLevel(level))
}
