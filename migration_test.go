package migration

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.uber.org/zap"
)

// To run these test you must have a mysql instance and a postgres instance running,
// or comment out all databases except Sqlite.
// Run docker-compose up to start mysql and postgres.

func TestService_Migrate(t *testing.T) {
	connectionStrings := map[SqlDialect]string{
		MySql:      "mig:mig@tcp(127.0.0.1:3306)/mig?parseTime=true",
		PostGreSQL: "postgresql://mig:mig@127.0.0.1:5432/mig",
		Sqlite:     "mig.db",
	}

	for d, c := range connectionStrings {
		t.Run(fmt.Sprintf("[%s] %s", d, "Init - ok"), func(t *testing.T) {
			s, err := New(d, c, zap.NewNop())
			if err != nil {
				assert.Fail(t, "failed to connect to database")
			}

			dropTables(s, d)

			s.WithFolder("test/init")
			err = s.Migrate()
			assert.NoError(t, err)

			tables, err := getTableNames(s, d)
			if err != nil {
				assert.Fail(t, "could not query for tables")
			}
			expectedTables := []string{"migration", "migration_lock", "test"}
			assert.True(t, reflect.DeepEqual(tables, expectedTables), "%v should be %v", tables, expectedTables)

		})

		t.Run(fmt.Sprintf("[%s] %s", d, "Init again - ok"), func(t *testing.T) {
			s, err := New(d, c, zap.NewNop())
			if err != nil {
				assert.Fail(t, "failed to connect to database")
			}

			s.WithFolder("test/init")
			err = s.Migrate()
			assert.NoError(t, err)

			tables, err := getTableNames(s, d)
			if err != nil {
				assert.Fail(t, "could not query for tables")
			}
			assert.True(t, reflect.DeepEqual(tables, []string{"migration", "migration_lock", "test"}))
		})

		t.Run(fmt.Sprintf("[%s] %s", d, "Different init - fail"), func(t *testing.T) {
			s, err := New(d, c, zap.NewNop())
			if err != nil {
				assert.Fail(t, "failed to connect to database")
			}

			s.WithFolder("test/diff-init")
			err = s.Migrate()
			assert.Error(t, err)

			tables, err := getTableNames(s, d)
			if err != nil {
				assert.Fail(t, "could not query for tables")
			}
			assert.True(t, reflect.DeepEqual(tables, []string{"migration", "migration_lock", "test"}))
		})

		t.Run(fmt.Sprintf("[%s] %s", d, "Multiple files - ok"), func(t *testing.T) {
			s, err := New(d, c, zap.NewNop())
			if err != nil {
				assert.Fail(t, "failed to connect to database")
			}

			s.WithFolder("test/init")
			err = s.Migrate()
			assert.NoError(t, err)

			tables, err := getTableNames(s, d)
			if err != nil {
				assert.Fail(t, "could not query for tables")
			}
			assert.True(t, reflect.DeepEqual(tables, []string{"migration", "migration_lock", "test"}))
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
		res, err := s.db.Query("select name from sqlite_master where type='table'")
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
