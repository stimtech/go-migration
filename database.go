package migration

import (
	"database/sql"
	"fmt"

	"go.uber.org/zap"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v4/stdlib"
)

type SqlDialect string

const (
	MySql      = SqlDialect("mysql")
	PostGreSQL = SqlDialect("pgx")
)

type Service struct {
	logger             *zap.Logger
	db                 *sql.DB
	migrationTable     string
	migrationLockTable string
	migrationFolder    string
	lockTimeoutMinutes int
}

// New returns a new Database instance
func New(dialect SqlDialect, connectionString string, logger *zap.Logger) (*Service, error) {
	db, err := sql.Open(string(dialect), connectionString)

	if err != nil {
		return nil, fmt.Errorf("failed to create database connection for migration: %w", err)
	}

	return &Service{
		logger:             logger,
		db:                 db,
		migrationTable:     "migration",
		migrationLockTable: "migration_lock",
		migrationFolder:    "db/migrations",
		lockTimeoutMinutes: 15,
	}, nil
}

// WithTableName changes the name of the migration table from 'migration'
func (s *Service) WithTableName(name string) {
	s.migrationTable = name
}

// WithLockTableName changes the name of the migration lock table from 'migration_lock'
func (s *Service) WithLockTableName(name string) {
	s.migrationLockTable = name
}

// WithFolder changes the location of migration sql files from 'db/migrations'
func (s *Service) WithFolder(name string) {
	s.migrationFolder = name
}

// WithLockTimeoutMinutes changes the lock timout from 15 minutes
func (s *Service) WithLockTimeoutMinutes(minutes int) {
	s.lockTimeoutMinutes = minutes
}
