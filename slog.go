package migration

import (
	"log/slog"
)

// SlogOption is used to define a custom slog logger to be used in the go-migration lib.
type SlogOption struct {
	Logger *slog.Logger
}

func (o SlogOption) apply(service *Service) {
	service.logger = slogLogger{logger: o.Logger}
}

type slogLogger struct {
	logger *slog.Logger
}

// Info logs a message at info level.
func (l slogLogger) Info(msg string) {
	l.logger.Info(msg)
}

// Warn logs a message with at warn level.
func (l slogLogger) Warn(msg string) {
	l.logger.Warn(msg)
}
