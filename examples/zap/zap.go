package main

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/stimtech/go-migration/v2"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

// Simple example app with a zap logger
// An sqlite database will be created, and a table added.
// Run with 'go run zap.go'.
// Two migrations are present in db/migration and will be loaded into the db in alphabetical order.

func main() {
	logger, _ := zap.NewProduction()

	// connect to a new or existing sqlite datasource
	db, err := sql.Open("sqlite3", "db.sqlite")
	if err != nil {
		logger.Fatal("failed to create datasource", zap.Error(err))
	}

	// run migration
	m := migration.New(db, migration.ZapOption{Logger: logger})
	err = m.Migrate()
	if err != nil {
		logger.Fatal("failed to run database migration", zap.Error(err))
	}

	// list all tables in the database
	res, err := db.Query("select name from sqlite_master where type='table' order by name")
	if err != nil {
		logger.Fatal("failed to query database", zap.Error(err))
	}

	var tables []string
	for res.Next() {
		var tn string
		if err := res.Scan(&tn); err != nil {
			logger.Fatal("failed to scan row", zap.Error(err))
		}
		tables = append(tables, tn)
	}

	logger.Info(fmt.Sprintf("Tables in database: %s", strings.Join(tables, ", ")))
}
