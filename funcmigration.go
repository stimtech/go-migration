package migration

import "database/sql"

// FuncMigration can be implemented by apps relying on go-migration to allow for
// code based migrations. If a call to Apply returns with a non-nil error, the
// migration process will be interrupted and a rollback will occur.
type FuncMigration interface {
	// Apply implementations should perform any arbitrary implementation function
	Apply(db *sql.Tx) error

	FileName() string
}
