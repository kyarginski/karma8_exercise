package postgres

import (
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"karma8/internal/app/repository"

	"github.com/stretchr/testify/require"
)

type TestDatabase struct {
	containerDatabase *TestContainerDatabase

	db *sql.DB
}

func NewTestDatabase(t *testing.T) (*TestDatabase, error) {
	containerDB := newTestContainerDatabase(t)
	// connectString := "host=localhost port=5432 dbname=postgres user=postgres password=postgres client_encoding=UTF8 sslmode=disable"
	connectString := containerDB.ConnectionString(t)

	storage, err := repository.New(connectString)
	if err != nil {
		return nil, err
	}
	testDB := storage.GetDB()
	testDatabase := &TestDatabase{
		containerDatabase: containerDB,
		db:                testDB,
	}

	err = testDatabase.prepareTestDBData(t)
	if err != nil {
		return nil, err
	}

	return testDatabase, nil
}

func (db *TestDatabase) DB() *sql.DB {
	return db.db
}

func (db *TestDatabase) Close(t *testing.T) {
	err := db.db.Close()
	require.NoError(t, err)
	db.containerDatabase.Close(t)
}

func (db *TestDatabase) readSQLFilesFromDir(t *testing.T) ([]string, error) {
	t.Helper()

	var sqlFiles []string

	path := db.readFixture(t)

	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if strings.HasSuffix(file.Name(), ".sql") {
			sqlFiles = append(sqlFiles, filepath.Join(path, file.Name()))
		}
	}

	sort.Strings(sqlFiles)

	return sqlFiles, nil
}

func (db *TestDatabase) executeSQLFile(t *testing.T, filePath string) error {
	t.Helper()

	sqlContent, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Split SQL script by semicolons.
	queries := strings.Split(string(sqlContent), ";")

	for _, query := range queries {
		if strings.TrimSpace(query) == "" {
			continue
		}
		_, err = db.db.Exec(query)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *TestDatabase) readFixture(t *testing.T) string {
	t.Helper()

	_, curFile, _, ok := runtime.Caller(0)
	require.True(t, ok)

	path := filepath.Join(filepath.Dir(curFile), "testdata")

	return path
}

func (db *TestDatabase) prepareTestDBData(t *testing.T) error {
	t.Helper()

	sqlFiles, err := db.readSQLFilesFromDir(t)
	if err != nil {
		return err
	}

	for _, sqlFile := range sqlFiles {
		err = db.executeSQLFile(t, sqlFile)
		if err != nil {
			return err
		}
	}
	if len(sqlFiles) != 0 {
		t.Logf("Test DB data is ready after executing %d SQL files", len(sqlFiles))
	}

	return nil
}

func (db *TestDatabase) ConnectString(t *testing.T) string {
	return db.containerDatabase.ConnectionString(t)
}
