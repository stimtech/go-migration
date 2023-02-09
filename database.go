package migration

import (
	"database/sql"
	"io/fs"
	"log"
	"os"
)

// Service is the db migration service.
type Service struct {
	logger             Logger
	db                 *sql.DB
	migrationTable     string
	migrationLockTable string
	migrationFolder    string
	lockTimeoutMinutes int
	fs                 fs.FS
}

// New returns a new Database instance.
func New(db *sql.DB, opts ...Option) *Service {
	s := &Service{
		db:                 db,
		logger:             defaultLogger{logger: log.New(os.Stdout, "go-migration: ", log.LstdFlags)},
		migrationTable:     "migration",
		migrationLockTable: "migration_lock",
		migrationFolder:    "db/migrations",
		lockTimeoutMinutes: 15,
		fs:                 os.DirFS("."),
	}
	for _, o := range opts {
		o.apply(s)
	}

	return s
}

// Logger is used to implement different logging solutions.
type Logger interface {
	Info(string)
	Warn(string)
}

// Option is used to configure go-migration in different ways. Please refer to the examples.
type Option interface {
	apply(service *Service)
}

// Config holds migration configuration parameters.
type Config struct {
	// TableName specifies the name of the table that keeps track of which migrations have been applied.
	// Defaults to "migration".
	TableName string

	// LockTableName specifies the name of the table that makes sure only one instance of go-migration runs at the
	// same time on the same database.
	// Defaults to "migration_lock".
	LockTableName string

	// MigrationFolder specifies the location of migration sql files.
	// Defaults to "db/migrations".
	MigrationFolder string

	// LockTimeoutMinutes specifies the lock timeout in minutes.
	// Defaults to 15.
	LockTimeoutMinutes int
}

func (c Config) apply(service *Service) {
	if c.TableName != "" {
		service.migrationTable = c.TableName
	}
	if c.LockTableName != "" {
		service.migrationLockTable = c.LockTableName
	}
	if c.MigrationFolder != "" {
		service.migrationFolder = c.MigrationFolder
	}
	if c.LockTimeoutMinutes > 0 {
		service.lockTimeoutMinutes = c.LockTimeoutMinutes
	}
}

// FSOption makes migration use a specific FileSystem, instead of the default.
// useful with embed, for example.
type FSOption struct {
	FileSystem fs.FS
}

func (o FSOption) apply(service *Service) {
	service.fs = o.FileSystem
}
