package migration

import (
	"testing"

	"go.uber.org/zap"

	"github.com/stretchr/testify/assert"
)

func TestService_WithFolder(t *testing.T) {
	s := New(nil, zap.NewNop(), Config{
		MigrationFolder: "test-name",
	})
	assert.Equal(t, &Service{
		logger:             s.logger,
		migrationTable:     "migration",
		migrationLockTable: "migration_lock",
		migrationFolder:    "test-name",
		lockTimeoutMinutes: 15,
	}, s)
}

func TestService_WithLockTableName(t *testing.T) {
	s := New(nil, zap.NewNop(), Config{
		LockTableName: "test-name",
	})
	assert.Equal(t, &Service{
		logger:             s.logger,
		migrationTable:     "migration",
		migrationLockTable: "test-name",
		migrationFolder:    "db/migrations",
		lockTimeoutMinutes: 15,
	}, s)
}

func TestService_WithLockTimeoutMinutes(t *testing.T) {
	s := New(nil, zap.NewNop(), Config{
		LockTimeoutMinutes: 20,
	})
	assert.Equal(t, &Service{
		logger:             s.logger,
		migrationTable:     "migration",
		migrationLockTable: "migration_lock",
		migrationFolder:    "db/migrations",
		lockTimeoutMinutes: 20,
	}, s)
}

func TestService_WithTableName(t *testing.T) {
	s := New(nil, zap.NewNop(), Config{
		TableName: "test-name",
	})
	assert.Equal(t, &Service{
		logger:             s.logger,
		migrationTable:     "test-name",
		migrationLockTable: "migration_lock",
		migrationFolder:    "db/migrations",
		lockTimeoutMinutes: 15,
	}, s)
}
