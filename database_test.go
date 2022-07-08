package migration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestService_WithFolder(t *testing.T) {
	s := &Service{}
	got := s.WithFolder("test-name")
	assert.Equal(t, "test-name", got.migrationFolder)
	assert.Equal(t, s, got)
}

func TestService_WithLockTableName(t *testing.T) {
	s := &Service{}
	got := s.WithLockTableName("test-name")
	assert.Equal(t, "test-name", got.migrationLockTable)
	assert.Equal(t, s, got)
}

func TestService_WithLockTimeoutMinutes(t *testing.T) {
	s := &Service{}
	got := s.WithLockTimeoutMinutes(42)
	assert.Equal(t, 42, got.lockTimeoutMinutes)
	assert.Equal(t, s, got)
}

func TestService_WithTableName(t *testing.T) {
	s := &Service{}
	got := s.WithTableName("test-name")
	assert.Equal(t, "test-name", got.migrationTable)
	assert.Equal(t, s, got)
}
