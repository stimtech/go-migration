package migrations

import (
	"database/sql"
	"fmt"
)

type Second struct {
	Name string
}

func (t *Second) Apply(tx *sql.Tx) error {
	if _, err := tx.Exec(`insert into my_table ("id") values (1);`); err != nil {
		return fmt.Errorf("failed to insert entry into my_table: %w", err)
	}

	if _, err := tx.Exec(`create table my_second_table (id int primary key, name varchar(128));`); err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	return nil
}

func (t *Second) Filename() string {
	return t.Name
}
