package migration

import (
	"database/sql"

	"go.uber.org/zap"
)

// Service is the db migration service
type Service struct {
	logger             *zap.Logger
	db                 *sql.DB
	migrationTable     string
	migrationLockTable string
	migrationFolder    string
	lockTimeoutMinutes int
}

// New returns a new Database instance.
func New(db *sql.DB, logger *zap.Logger, config Config) *Service {
	config = config.replaceEmptiesWithDefaults()
	return &Service{
		logger:             logger.Named("go-migration"),
		db:                 db,
		migrationTable:     config.TableName,
		migrationLockTable: config.LockTableName,
		migrationFolder:    config.MigrationFolder,
		lockTimeoutMinutes: config.LockTimeoutMinutes,
	}
}

// Config holds migration configuration parameters
type Config struct {
	// TableName specifies the name of the table that keeps track of which migrations have been applied.
	// Defaults to "migration"
	TableName string

	// LockTableName specifies the name of the table that makes sure only one instance of go-migration runs at the
	// same time on the same database.
	// Defaults to "migration_lock"
	LockTableName string

	// MigrationFolder specifies the location of migration sql files.
	// Defaults to "db/migrations"
	MigrationFolder string

	// LockTimeoutMinutes specifies the lock timeout in minutes.
	// Defaults to 15
	LockTimeoutMinutes int
}

func (c Config) replaceEmptiesWithDefaults() Config {
	if c.TableName == "" {
		c.TableName = "migration"
	}
	if c.LockTableName == "" {
		c.LockTableName = "migration_lock"
	}
	if c.MigrationFolder == "" {
		c.MigrationFolder = "db/migrations"
	}
	if c.LockTimeoutMinutes <= 0 {
		c.LockTimeoutMinutes = 15
	}
	return c
}
