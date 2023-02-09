package migration

import (
	"go.uber.org/zap"
)

// ZapOption is used to define a custom zap logger to be used in the go-migration lib.
type ZapOption struct {
	Logger *zap.Logger
}

func (z ZapOption) apply(service *Service) {
	service.logger = zapLogger{logger: z.Logger}
}

type zapLogger struct {
	logger *zap.Logger
}

// Info logs a message at info level.
func (l zapLogger) Info(msg string) {
	l.logger.Info(msg)
}

// Warn logs a message with at warn level.
func (l zapLogger) Warn(msg string) {
	l.logger.Warn(msg)
}
