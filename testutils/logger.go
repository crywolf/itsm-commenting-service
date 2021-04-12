package testutils

import "go.uber.org/zap"

// NewTestLogger returns new logger with level set to ErrorLevel
func NewTestLogger() *zap.Logger {
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	return logger
}
