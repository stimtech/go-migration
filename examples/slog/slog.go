package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"strings"

	"github.com/stimtech/go-migration/v2"

	_ "github.com/mattn/go-sqlite3"
)

// Simple example app with a zap logger
// An sqlite database will be created, and a table added.
// Run with 'go run basic.go'.
// Two migrations are present in db/migration and will be loaded into the db in alphabetical order.

func main() {
	logger := slog.Default()

	// connect to a new or existing sqlite datasource
	db, err := sql.Open("sqlite3", "db.sqlite")
	if err != nil {
		logger.Error("failed to create datasource", slog.Any("error", err))
		return
	}

	// run migration
	m := migration.New(db, migration.SlogOption{Logger: logger})

	err = m.Migrate()
	if err != nil {
		logger.Error("failed to run database migration", slog.Any("error", err))
		return
	}

	// list all tables in the database
	res, err := db.Query("select name from sqlite_master where type='table' order by name")
	if err != nil {
		logger.Error("failed to query database", slog.Any("error", err))
		return
	}

	var tables []string

	for res.Next() {
		var tn string
		if err := res.Scan(&tn); err != nil {
			logger.Error("failed to scan row", slog.Any("error", err))
		}

		tables = append(tables, tn)
	}

	logger.Info(fmt.Sprintf("Tables in database: %s", strings.Join(tables, ", ")))
}
