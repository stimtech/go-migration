package migration

import "database/sql"

// FuncMigration can be implemented by apps relying on go-migration to allow for
// code based migrations. If a call to Apply returns with a non-nil error, the
// migration process will be interrupted and a rollback will occur.
type FuncMigration interface {
	// Apply should perform the migration. Implementations should not commit
	// nor rollback the transaction.
	Apply(db *sql.Tx) error

	// Filename should declare the file in the migrations dir in which the
	// migration is implemented.
	Filename() string
}
