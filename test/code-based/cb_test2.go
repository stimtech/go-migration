// nolint:revive
// No docs for test migration implementations.
package code_based

import (
	"database/sql"
	"fmt"
)

type CBTest2 struct {
	Name string
}

func (t *CBTest2) Apply(tx *sql.Tx) error {
	if _, err := tx.Exec(`insert into cb_1_initial ("id") values ("grodanboll");`); err != nil {
		return fmt.Errorf("failed to insert entry into cb_1_initial table: %w", err)
	}

	if _, err := tx.Exec(`create table "cb_2_from_code" (name varchar(100) not null);`); err != nil {
		return fmt.Errorf("failed to modify schema: %w", err)
	}

	return nil
}

func (t *CBTest2) Filename() string {
	return t.Name
}
