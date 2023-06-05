package code_based

import (
	"database/sql"
	"fmt"
)

type CBFailTest2 struct {
	Name string
}

func (t *CBFailTest2) Apply(tx *sql.Tx) error {
	// Valid statement.
	if _, err := tx.Exec(`insert into cb_2_initial ("id") values ("rollmeback");`); err != nil {
		return fmt.Errorf("failed to insert entry into cb_2_initial table: %w", err)
	}

	// Bad statement, previous change should be rolled back.
	if _, err := tx.Exec(`insert into non_existing ("fugazi") values ("stuff");`); err != nil {
		return fmt.Errorf("failed to insert entry into cb_2_initial table: %w", err)
	}

	return nil
}

func (t *CBFailTest2) Filename() string {
	return t.Name
}
