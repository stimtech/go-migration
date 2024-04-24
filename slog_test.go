package migration

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SlogOption_apply(t *testing.T) {
	l := slog.Default()
	s := &Service{}
	SlogOption{Logger: l}.apply(s)
	assert.Equal(t, l, s.logger.(slogLogger).logger)
}

func Test_slogLogger_Info(t *testing.T) {
	mockLog := &mockSlogHandler{}

	l := slogLogger{logger: slog.New(mockLog)}
	l.Info("test message")

	if assert.Len(t, mockLog.Infos, 1) {
		assert.Equal(t, "test message", mockLog.Infos[0])
	}
}

func Test_slogLogger_Warn(t *testing.T) {
	mockLog := &mockSlogHandler{}

	l := slogLogger{logger: slog.New(mockLog)}
	l.Warn("test message")
	assert.Len(t, mockLog.Warns, 1)
	assert.Equal(t, "test message", mockLog.Warns[0])
}

type mockSlogHandler struct {
	Infos []string
	Warns []string
}

// Enabled is a mock method.
func (h *mockSlogHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

// WithAttrs is a mock method.
func (h *mockSlogHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

// WithGroup is a mock method.
func (h *mockSlogHandler) WithGroup(string) slog.Handler {
	return h
}

// Handle adds slog.Records to the error and infos array. Used in tests.
func (h *mockSlogHandler) Handle(_ context.Context, r slog.Record) error {
	if r.Level == slog.LevelWarn {
		h.Warns = append(h.Warns, r.Message)
	}

	if r.Level == slog.LevelInfo {
		h.Infos = append(h.Infos, r.Message)
	}

	return nil
}
