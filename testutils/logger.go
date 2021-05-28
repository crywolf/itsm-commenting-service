package testutils

import "go.uber.org/zap"

// NewTestLogger returns new logger with level set to FatalLevel and a reference to  it's config
func NewTestLogger() (*zap.Logger, *zap.Config) {
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(zap.FatalLevel)
	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	return logger, &cfg
}
