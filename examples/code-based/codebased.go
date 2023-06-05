package main

import (
	"database/sql"
	"fmt"
	"github.com/stimtech/go-migration/v2/examples/code-based/db/migrations"
	"log"
	"strings"

	"github.com/stimtech/go-migration/v2"

	_ "github.com/mattn/go-sqlite3"
)

// Simple example app to showcase use of go based migrations.
// Run with 'go run code-based.go'.
// Two migrations are present in db/migration and will be loaded into the db in
// alphabetical order.
func main() {
	// connect to a new or existing sqlite datasource
	db, err := sql.Open("sqlite3", "db.sqlite")
	if err != nil {
		log.Fatal("Failed to create datasource.", err)
	}

	// Code based migrations that are included in the migrations dir need to be
	// explicitly declared. Be careful about providing the correct
	// implementation / name combination as this is not run- nor compile time
	// checked for accuracy.
	m := migration.New(db, migration.FuncMigrationOption{
		Migration: &migrations.Second{
			// The name here needs to match a filename in the migrations' dir.
			Name: "2023-06-05-second.go",
		},
	})
	err = m.Migrate()
	if err != nil {
		log.Fatal("Failed to run database migrations.", err)
	}

	// list all tables in the database
	res, err := db.Query("select name from sqlite_master where type='table' order by name")
	if err != nil {
		log.Fatal("Failed to query database.", err)
	}

	var tables []string
	for res.Next() {
		var tn string
		if err := res.Scan(&tn); err != nil {
			log.Fatal("Failed to scan row.", err)
		}
		tables = append(tables, tn)
	}

	log.Println(fmt.Sprintf("Tables in database: %s", strings.Join(tables, ", ")))
}
