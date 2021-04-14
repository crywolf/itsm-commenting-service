package testutils

import "go.uber.org/zap"

// NewTestLogger returns new logger with level set to FatalLevel
func NewTestLogger() *zap.Logger {
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(zap.FatalLevel)
	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	return logger
}
