package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Settings contains logging related parameters
type Settings struct {
	Level         string
	Output        []string
	LogCaller     bool
	LogStacktrace bool
}

// Logger is a global that is used for logging
var Logger *zap.Logger

// Init creates logger and configures it
func Init(logSettings Settings) error {
	// level defines current logging level: debug, info, warn, etc
	level := zapcore.InfoLevel
	err := level.Set(logSettings.Level)
	if err != nil {
		Logger.Error("Could not set log level, using info as default", zap.Error(err), zap.String("level", logSettings.Level))
		return err
	}

	zapConfig := zap.NewProductionConfig()
	zapConfig.OutputPaths = logSettings.Output
	zapConfig.Level = zap.NewAtomicLevelAt(level)
	zapConfig.DisableCaller = !logSettings.LogCaller
	zapConfig.DisableStacktrace = !logSettings.LogStacktrace
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	Logger, err = zapConfig.Build()
	if err == nil {
		Logger = Logger.With(serviceNameField)
	}
	return err
}

var serviceNameField = zap.String("service", "reactions")
