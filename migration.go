package migration

import (
	"bytes"
	"crypto/md5" //nolint:gosec
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"sort"
	"strings"
	"time"
)

type migration struct {
	ID       string    // unique ID for the migration. Also name of file
	Date     time.Time // date and time the migration was applied, default current_timestamp
	Checksum string    // makes sure that migrations do not change over time
}

// Migrate applies all non applied migrations in the migration folder to the database, in alphabetical order.
// It also checks that previously applied migrations have not changed using md5 checksums.
func (s *Service) Migrate() error {
	err := s.createMigrationTables()
	if err != nil {
		return fmt.Errorf("failed to create migration tables: %w", err)
	}

	locked, release := s.lock()
	defer release()

	if !locked {
		return errors.New("migration already in progress. failed to get lock")
	}

	appliedMigs, err := s.fetchAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to fetch applied migrations: %w", err)
	}

	availableMigs, err := s.listMigrations()
	if err != nil {
		return fmt.Errorf("failed to list available migrations: %w", err)
	}

	sort.Strings(availableMigs)

	for _, mig := range availableMigs {
		chkSum, applied := appliedMigs[mig]

		if !applied {
			funcMigration, err := s.shouldApplyFuncMigration(mig)
			if err != nil {
				return fmt.Errorf("failed to determine if func migration should be applied: %w", err)
			}

			if funcMigration != nil {
				// Code based migration not yet applied was found.
				if err := s.applyFuncMigration(funcMigration); err != nil {
					return fmt.Errorf("failed to apply func migration %s: %w", mig, err)
				}

				continue
			}

			// SQL based migration not yet applied was found.
			if err := s.applySQLMigration(mig); err != nil {
				return fmt.Errorf("failed to apply migration %s: %w", mig, err)
			}
		} else {
			c, err := s.fileHash(fmt.Sprintf("%s/%s", s.migrationFolder, mig))
			if err != nil {
				return fmt.Errorf("failed to get checksum for file %s: %w", mig, err)
			}
			if c != chkSum {
				return fmt.Errorf("file %s has been updated since it was migrated", mig)
			}
		}
	}

	return nil
}

func (s *Service) shouldApplyFuncMigration(name string) (FuncMigration, error) {
	if !strings.HasSuffix(name, ".go") {
		return nil, nil
	}

	fm, exists := s.funcMigrations[name]
	if !exists {
		return nil, fmt.Errorf("failed to find supplied func migration implementation with filename %s", name)
	}

	if fm.Filename() != name {
		return nil, fmt.Errorf("declared filename of "+
			"funcmigration does not match filename of "+
			"migration operated on, expected %s got %s",
			name,
			fm.Filename(),
		)
	}

	return fm, nil
}

func (s *Service) createMigrationTables() error {
	_, err := s.db.Exec(fmt.Sprintf(`create table if not exists %s (
		id varchar(255) primary key,
		date timestamp default current_timestamp,
		checksum varchar(255));`,
		s.migrationTable))
	if err != nil {
		return err
	}

	_, err = s.db.Exec(fmt.Sprintf(`create table if not exists %s (
		id integer primary key,
		created_at timestamp default current_timestamp);`,
		s.migrationLockTable))

	return err
}

func (s *Service) lock() (bool, func()) {
	release := func() {
		_, _ = s.db.Exec(fmt.Sprintf("delete from %s", s.migrationLockTable))
	}

	for i := 0; i < 12; i++ {
		_, err := s.db.Exec(fmt.Sprintf("insert into %s(id) values(1)", s.migrationLockTable))
		if err == nil {
			return true, release
		}

		s.logger.Info("waiting for migration lock")
		time.Sleep(5 * time.Second)

		_, _ = s.db.Exec(
			fmt.Sprintf("delete from %s where created_at < timestampadd(minute, %d, current_timestamp)",
				s.migrationLockTable, -s.lockTimeoutMinutes))
	}

	return false, func() {}
}

func (s *Service) fetchAppliedMigrations() (map[string]string, error) {
	var mig migration

	rows, err := s.db.Query(fmt.Sprintf("select * from %s", s.migrationTable))
	if err != nil {
		return nil, err
	}

	migMap := make(map[string]string)

	for rows.Next() {
		err = rows.Scan(&mig.ID, &mig.Date, &mig.Checksum)
		migMap[mig.ID] = mig.Checksum
	}

	if err != nil {
		return nil, err
	}

	return migMap, nil
}

func (s *Service) listMigrations() ([]string, error) {
	files, err := fs.ReadDir(s.fs, s.migrationFolder)
	if err != nil {
		return nil, err
	}

	var fileNames []string

	for _, file := range files {
		if !file.IsDir() {
			fileNames = append(fileNames, file.Name())
		}
	}

	return fileNames, nil
}

func (s *Service) applySQLMigration(mig string) error {
	c, err := s.fileHash(fmt.Sprintf("%s/%s", s.migrationFolder, mig))
	if err != nil {
		return fmt.Errorf("failed to get checksum for file %s: %w", mig, err)
	}

	file, err := fs.ReadFile(s.fs, fmt.Sprintf("%s/%s", s.migrationFolder, mig))

	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", mig, err)
	}

	requests := strings.Split(string(file), ";")

	s.logger.Info(fmt.Sprintf("applying migration: %s", mig))

	// MySQL transactions will not work with ALTER TABLE and other DDL statements. See this post for more details:
	// https://stackoverflow.com/questions/22806261/can-i-use-transactions-with-alter-table
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	for _, request := range requests {
		if strings.Trim(request, " \n\r") == "" {
			continue
		}

		_, err = tx.Exec(request)
		if err != nil {
			if err := tx.Rollback(); err != nil {
				s.logger.Warn("rollback failed")
			}

			return fmt.Errorf("failing statement [%s]: %w", request, err)
		}
	}

	if err = s.insertCompletedMigration(tx, c, mig); err != nil {
		return fmt.Errorf("failed to insert migration: %w", err)
	}

	if err != nil {
		if err := tx.Rollback(); err != nil {
			s.logger.Warn("rollback failed")
		}

		return err
	}

	return tx.Commit()
}

func (s *Service) applyFuncMigration(fm FuncMigration) error {
	s.logger.Info(fmt.Sprintf("applying migration: %s", fm.Filename()))

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err := fm.Apply(tx); err != nil {
		if err := tx.Rollback(); err != nil {
			return fmt.Errorf("failed to rollback failed migration %w", err)
		}

		return fmt.Errorf("failed to apply migration: %w", err)
	}

	// Checksum of func migrations are only based on the filename of the
	// migration. This is to prevent issues where import names may change due
	// to lib updates, which would otherwise break the hashing contract.
	checksum, err := s.checksum(bytes.NewReader([]byte(fm.Filename())))
	if err != nil {
		return fmt.Errorf("failed to create checksum for migration: %w", err)
	}

	if err := s.insertCompletedMigration(tx, checksum, fm.Filename()); err != nil {
		return fmt.Errorf("failed to insert migration: %w", err)
	}

	return tx.Commit()
}

func (s *Service) insertCompletedMigration(tx *sql.Tx, checksum, filename string) error {
	if _, err := tx.Exec(
		fmt.Sprintf(
			`insert into %s (id, checksum) values ('%s', '%s')`,
			s.migrationTable,
			filename,
			checksum,
		),
	); err != nil {
		return fmt.Errorf("failed to insert applied migration into migrations table: %w", err)
	}

	return nil
}

func (s *Service) fileHash(filename string) (string, error) {
	input, err := s.fs.Open(filename)
	if err != nil {
		return "", err
	}

	return s.checksum(input)
}

func (s *Service) checksum(input io.Reader) (string, error) {
	hash := md5.New() // nolint:gosec
	if _, err := io.Copy(hash, input); err != nil {
		return "", err
	}

	sum := hash.Sum(nil)

	return hex.EncodeToString(sum), nil
}
