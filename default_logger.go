package migration

import (
	"log"
)

// LoggerOption is used to define a custom logger to be used in the go-migration lib.
type LoggerOption struct {
	Logger *log.Logger
}

func (l LoggerOption) apply(service *Service) {
	service.logger = defaultLogger{logger: l.Logger}
}

type defaultLogger struct {
	logger *log.Logger
}

// Info prints a message
func (l defaultLogger) Info(msg string) {
	l.logger.Println(msg)
}

// Warn prints a message prefixed with 'warning: '
func (l defaultLogger) Warn(msg string) {
	l.logger.Print("warning: %s \n", msg)
}
