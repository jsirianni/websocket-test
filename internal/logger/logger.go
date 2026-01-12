package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a new zap logger configured for JSON output to stdout
// with a common time format (RFC3339).
func New() (*zap.Logger, error) {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	config.EncodeLevel = zapcore.LowercaseLevelEncoder

	encoder := zapcore.NewJSONEncoder(config)
	core := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zapcore.InfoLevel)

	return zap.New(core), nil
}

// MustNew creates a new zap logger and panics if initialization fails.
// This is useful for initialization at program startup.
func MustNew() *zap.Logger {
	logger, err := New()
	if err != nil {
		panic(err)
	}
	return logger
}

