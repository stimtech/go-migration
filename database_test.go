package migration

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestService_WithFolder(t *testing.T) {
	s := New(nil, Config{
		MigrationFolder: "test-name",
	})
	assert.Equal(t, &Service{
		logger:             s.logger,
		migrationTable:     "migration",
		migrationLockTable: "migration_lock",
		migrationFolder:    "test-name",
		lockTimeoutMinutes: 15,
		fs:                 os.DirFS("."),
	}, s)
}

func TestService_WithLockTableName(t *testing.T) {
	s := New(nil, Config{
		LockTableName: "test-name",
	})
	assert.Equal(t, &Service{
		logger:             s.logger,
		migrationTable:     "migration",
		migrationLockTable: "test-name",
		migrationFolder:    "db/migrations",
		lockTimeoutMinutes: 15,
		fs:                 os.DirFS("."),
	}, s)
}

func TestService_WithLockTimeoutMinutes(t *testing.T) {
	s := New(nil, Config{
		LockTimeoutMinutes: 20,
	})
	assert.Equal(t, &Service{
		logger:             s.logger,
		migrationTable:     "migration",
		migrationLockTable: "migration_lock",
		migrationFolder:    "db/migrations",
		lockTimeoutMinutes: 20,
		fs:                 os.DirFS("."),
	}, s)
}

func TestService_WithTableName(t *testing.T) {
	s := New(nil, Config{
		TableName: "test-name",
	})
	assert.Equal(t, &Service{
		logger:             s.logger,
		migrationTable:     "test-name",
		migrationLockTable: "migration_lock",
		migrationFolder:    "db/migrations",
		lockTimeoutMinutes: 15,
		fs:                 os.DirFS("."),
	}, s)
}
