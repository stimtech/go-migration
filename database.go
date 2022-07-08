package migration

import (
	"database/sql"

	"go.uber.org/zap"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v4/stdlib"
	_ "github.com/mattn/go-sqlite3"
)

type Service struct {
	logger             *zap.Logger
	db                 *sql.DB
	migrationTable     string
	migrationLockTable string
	migrationFolder    string
	lockTimeoutMinutes int
}

// New returns a new Database instance.
func New(db *sql.DB, logger *zap.Logger) *Service {
	// go-migrate needs two connections
	if db.Stats().MaxOpenConnections == 1 {
		db.SetMaxOpenConns(2)
	}
	return &Service{
		logger:             logger.Named("go-migration"),
		db:                 db,
		migrationTable:     "migration",
		migrationLockTable: "migration_lock",
		migrationFolder:    "db/migrations",
		lockTimeoutMinutes: 15,
	}
}

// WithTableName changes the name of the migration table from 'migration'
func (s *Service) WithTableName(name string) *Service {
	s.migrationTable = name
	return s
}

// WithLockTableName changes the name of the migration lock table from 'migration_lock'
func (s *Service) WithLockTableName(name string) *Service {
	s.migrationLockTable = name
	return s
}

// WithFolder changes the location of migration sql files from 'db/migrations'
func (s *Service) WithFolder(name string) *Service {
	s.migrationFolder = name
	return s
}

// WithLockTimeoutMinutes changes the lock timout from 15 minutes
func (s *Service) WithLockTimeoutMinutes(minutes int) *Service {
	s.lockTimeoutMinutes = minutes
	return s
}
