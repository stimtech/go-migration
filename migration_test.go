package migration

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.uber.org/zap"
)

type SqlDialect string

const (
	MySql      = SqlDialect("mysql")
	PostGreSQL = SqlDialect("pgx")
	Sqlite     = SqlDialect("sqlite3")
)

// To run these tests on mysql and postgres, uncomment the connectionStrings
// and start mysql and postgres using `docker-compose up`.

func TestService_Migrate(t *testing.T) {
	connectionStrings := map[SqlDialect]string{
		//MySql:      "mig:mig@tcp(127.0.0.1:3306)/mig?parseTime=true",
		//PostGreSQL: "postgresql://mig:mig@127.0.0.1:5432/mig",
		Sqlite: "mig.db",
	}

	for d, c := range connectionStrings {
		t.Run(fmt.Sprintf("[%s] %s", d, "Init - ok"), func(t *testing.T) {
			db, err := sql.Open(string(d), c)
			if !assert.NoError(t, err) {
				return
			}
			defer func() { _ = db.Close() }()

			s := New(db, zap.NewNop())

			dropTables(s, d)

			s.WithFolder("test/init")
			err = s.Migrate()
			assert.NoError(t, err)

			tables, err := getTableNames(s, d)
			if !assert.NoError(t, err, "could not query for tables") {
				return
			}
			assert.Equal(t, []string{"migration", "migration_lock", "test"}, tables)

		})

		t.Run(fmt.Sprintf("[%s] %s", d, "Init again - ok"), func(t *testing.T) {
			db, err := sql.Open(string(d), c)
			if !assert.NoError(t, err) {
				return
			}
			defer func() { _ = db.Close() }()

			s := New(db, zap.NewNop())

			s.WithFolder("test/init")
			err = s.Migrate()
			assert.NoError(t, err)

			tables, err := getTableNames(s, d)
			if !assert.NoError(t, err, "could not query for tables") {
				return
			}
			assert.Equal(t, []string{"migration", "migration_lock", "test"}, tables)
		})

		t.Run(fmt.Sprintf("[%s] %s", d, "Different init - fail"), func(t *testing.T) {
			db, err := sql.Open(string(d), c)
			if !assert.NoError(t, err) {
				return
			}
			defer func() { _ = db.Close() }()

			s := New(db, zap.NewNop())

			s.WithFolder("test/diff-init")
			err = s.Migrate()
			assert.Error(t, err)

			tables, err := getTableNames(s, d)
			if !assert.NoError(t, err, "could not query for tables") {
				return
			}
			assert.Equal(t, []string{"migration", "migration_lock", "test"}, tables)
		})

		t.Run(fmt.Sprintf("[%s] %s", d, "Multiple files - ok"), func(t *testing.T) {
			db, err := sql.Open(string(d), c)
			if !assert.NoError(t, err) {
				return
			}
			defer func() { _ = db.Close() }()

			s := New(db, zap.NewNop())

			s.WithFolder("test/multi")
			err = s.Migrate()
			assert.NoError(t, err)

			tables, err := getTableNames(s, d)
			if !assert.NoError(t, err, "could not query for tables") {
				return
			}
			assert.Equal(t, []string{"migration", "migration_lock", "multi", "multi2", "test"}, tables)
		})

		t.Run(fmt.Sprintf("[%s] %s", d, "Failing statement - fail"), func(t *testing.T) {
			db, err := sql.Open(string(d), c)
			if !assert.NoError(t, err) {
				return
			}
			defer func() { _ = db.Close() }()

			s := New(db, zap.NewNop())

			s.WithFolder("test/failing-stmt")
			err = s.Migrate()
			assert.Error(t, err)

			tables, err := getTableNames(s, d)
			if !assert.NoError(t, err, "could not query for tables") {
				return
			}
			if d == MySql { // MySql can't roll back DDL statements
				assert.Equal(t, []string{"migration", "migration_lock", "multi", "multi2", "should_rollback", "test"}, tables)
			} else {
				assert.Equal(t, []string{"migration", "migration_lock", "multi", "multi2", "test"}, tables)
			}
		})

		t.Run(fmt.Sprintf("[%s] %s", d, "Missing folder - fail"), func(t *testing.T) {
			db, err := sql.Open(string(d), c)
			if !assert.NoError(t, err) {
				return
			}
			defer func() { _ = db.Close() }()

			s := New(db, zap.NewNop())

			s.WithFolder("test/no-folder")
			err = s.Migrate()
			assert.Error(t, err)
			if assert.Error(t, err) {
				assert.True(t, strings.Contains(err.Error(), "failed to list available migrations"))
			}
		})

		t.Run(fmt.Sprintf("[%s] %s", d, "No read access to file - fail"), func(t *testing.T) {
			db, err := sql.Open(string(d), c)
			if !assert.NoError(t, err) {
				return
			}
			defer func() { _ = db.Close() }()

			s := New(db, zap.NewNop())
			s.WithFolder("test/no-access")

			_ = os.Chmod("test/no-access/no-access.sql", os.ModeExclusive)
			defer func() { _ = os.Chmod("test/no-access/no-access.sql", os.ModePerm) }()

			err = s.Migrate()
			if assert.Error(t, err) {
				assert.True(t, strings.Contains(err.Error(), "permission denied"))
			}

		})
	}
}

func getTableNames(s *Service, d SqlDialect) ([]string, error) {
	var tables []string
	switch d {
	case MySql:
		res, err := s.db.Query("show tables")
		if err != nil {
			return nil, err
		}

		for res.Next() {
			var tn string
			if err := res.Scan(&tn); err != nil {
				return nil, err
			}
			tables = append(tables, tn)
		}
	case PostGreSQL:
		res, err := s.db.Query("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' ORDER BY table_name")
		if err != nil {
			return nil, err
		}

		for res.Next() {
			var tn string
			if err := res.Scan(&tn); err != nil {
				return nil, err
			}
			tables = append(tables, tn)
		}
	case Sqlite:
		res, err := s.db.Query("select name from sqlite_master where type='table' order by name")
		if err != nil {
			return nil, err
		}

		for res.Next() {
			var tn string
			if err := res.Scan(&tn); err != nil {
				return nil, err
			}
			tables = append(tables, tn)
		}
	}

	return tables, nil
}

func dropTables(s *Service, d SqlDialect) {
	tables, _ := getTableNames(s, d)
	for _, t := range tables {
		_, _ = s.db.Exec("drop table if exists " + t)
	}
}
