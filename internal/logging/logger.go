package logging

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(level string) (*zap.Logger, error) {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = os.Getenv("JSEER_ENV")
	}
	isDev := env == "dev" || env == "development"
	cfg := zap.NewProductionConfig()
	if isDev {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.TimeKey = "ts"
		cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05")
		cfg.Encoding = "console"
	}
	switch level {
	case "debug":
		cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "warn":
		cfg.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		cfg.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	default:
		cfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}
	return cfg.Build()
}
