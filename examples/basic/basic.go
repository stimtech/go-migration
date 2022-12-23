package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/stimtech/go-migration/v2"

	_ "github.com/mattn/go-sqlite3"
)

// Simple example app with minimal configuration
// An sqlite database will be created, and a table added.
// Run with 'go run basic.go'.
// Two migrations are present in db/migration and will be loaded into the db in alphabetical order.

func main() {
	// connect to a new or existing sqlite datasource
	db, err := sql.Open("sqlite3", "db.sqlite")
	if err != nil {
		log.Fatal("failed to create datasource", err)
	}

	// run migration
	m := migration.New(db)
	err = m.Migrate()
	if err != nil {
		log.Fatal("failed to run database migration", err)
	}

	// list all tables in the database
	res, err := db.Query("select name from sqlite_master where type='table' order by name")
	if err != nil {
		log.Fatal("failed to query database", err)
	}

	var tables []string
	for res.Next() {
		var tn string
		if err := res.Scan(&tn); err != nil {
			log.Fatal("failed to scan row", err)
		}
		tables = append(tables, tn)
	}

	log.Println(fmt.Sprintf("Tables in database: %s", strings.Join(tables, ", ")))
}
