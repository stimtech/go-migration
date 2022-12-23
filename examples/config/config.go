package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/stimtech/go-migration/v2"

	_ "github.com/mattn/go-sqlite3"
)

// Simple example app with some configuration
// An sqlite database will be created, and a table added.
// Run with 'go run config.go'.
// Two migrations are present in 'sql' and will be loaded into the db in alphabetical order.

func main() {
	logger := log.New(os.Stdout, "config-example: ", log.LstdFlags)
	// connect to a new or existing sqlite datasource
	db, err := sql.Open("sqlite3", "db.sqlite")
	if err != nil {
		logger.Fatal("failed to create datasource", err)
	}

	// run migration
	m := migration.New(db,
		migration.LoggerOption{Logger: logger}, // custom logger
		migration.Config{
			MigrationFolder: "sql", // changes the folder where migration sql files are
			TableName:       "example_migrations",
			LockTableName:   "example_migrations_lock",
		},
	)
	err = m.Migrate()
	if err != nil {
		logger.Fatal("failed to run database migration", err)
	}

	// list all tables in the database
	res, err := db.Query("select name from sqlite_master where type='table' order by name")
	if err != nil {
		logger.Fatal("failed to query database", err)
	}

	var tables []string
	for res.Next() {
		var tn string
		if err := res.Scan(&tn); err != nil {
			log.Fatal("failed to scan row", err)
		}
		tables = append(tables, tn)
	}

	logger.Println(fmt.Sprintf("Tables in database: %s", strings.Join(tables, ", ")))
}
