package migration

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
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
		chkSum, ok := appliedMigs[mig]

		if !ok {
			err = s.applyMigration(mig)
			if err != nil {
				return fmt.Errorf("failed to apply migration %s: %w", mig, err)
			}
		} else {
			c, err := checkSum(fmt.Sprintf("%s/%s", s.migrationFolder, mig))

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
		_, _ = s.db.Exec(fmt.Sprintf("delete from %s where created_at < timestampadd(minute, %d, current_timestamp)", s.migrationLockTable, -s.lockTimeoutMinutes))
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
	files, err := ioutil.ReadDir(s.migrationFolder)
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

func (s *Service) applyMigration(mig string) error {
	c, err := checkSum(fmt.Sprintf("%s/%s", s.migrationFolder, mig))
	if err != nil {
		return fmt.Errorf("failed to get checksum for file %s: %w", mig, err)
	}

	file, err := ioutil.ReadFile(s.migrationFolder + "/" + mig)

	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", mig, err)
	}

	requests := strings.Split(string(file), ";")

	s.logger.Info("applying migration", zap.String("fileName", mig))

	// MySql transactions will not work with ALTER TABLE and other DDL statements. See this post for more details:
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
			s.logger.Error("failing statement", zap.String("sql", request))
			if err := tx.Rollback(); err != nil {
				s.logger.Warn("rollback failed")
			}
			return err
		}
	}
	_, err = tx.Exec(fmt.Sprintf(`insert into %s (id, checksum) values ('%s', '%s')`, s.migrationTable, mig, c))
	if err != nil {
		if err := tx.Rollback(); err != nil {
			s.logger.Warn("rollback failed")
		}
		return err
	}

	return tx.Commit()
}

func checkSum(filename string) (string, error) {
	input, err := os.Open(filename)
	if err != nil {
		return "", err
	}

	hash := md5.New()
	if _, err := io.Copy(hash, input); err != nil {
		return "", err
	}
	sum := hash.Sum(nil)
	return hex.EncodeToString(sum), nil
}
