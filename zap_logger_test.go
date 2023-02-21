package migration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func Test_ZapOption_apply(t *testing.T) {
	l := zap.NewNop()
	s := &Service{}
	ZapOption{Logger: l}.apply(s)
	assert.Equal(t, l, s.logger.(zapLogger).logger)
}

func Test_zapLogger_Info(t *testing.T) {
	observedZapCore, observedLogs := observer.New(zap.InfoLevel)
	observedLogger := zap.New(observedZapCore)

	l := zapLogger{logger: observedLogger}
	l.Info("test message")
	assert.Equal(t, 1, observedLogs.Len())
	assert.Equal(t, "test message", observedLogs.All()[0].Message)
	assert.Equal(t, zap.InfoLevel, observedLogs.All()[0].Level)
}

func Test_zapLogger_Warn(t *testing.T) {
	observedZapCore, observedLogs := observer.New(zap.InfoLevel)
	observedLogger := zap.New(observedZapCore)

	l := zapLogger{logger: observedLogger}
	l.Warn("test message")
	assert.Equal(t, 1, observedLogs.Len())
	assert.Equal(t, "test message", observedLogs.All()[0].Message)
	assert.Equal(t, zap.WarnLevel, observedLogs.All()[0].Level)
}
