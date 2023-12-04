package redis

import (
	"testing"

	"karma8/internal/app/repository"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

type TestRedis struct {
	containerDatabase *TestContainerRedis

	db *redis.Client
}

func NewTestRedis(t *testing.T) (*TestRedis, error) {
	containerDB := newTestContainerRedis(t)
	// connectString := "redis://localhost:6379?protocol=3"
	connectString := containerDB.ConnectionString(t)

	storage, err := repository.NewRedis(connectString, 0)
	if err != nil {
		return nil, err
	}
	testDB := storage.GetDB()
	testRedis := &TestRedis{
		containerDatabase: containerDB,
		db:                testDB,
	}

	return testRedis, nil
}

func (db *TestRedis) DB() *redis.Client {
	return db.db
}

func (db *TestRedis) Close(t *testing.T) {
	err := db.db.Close()
	require.NoError(t, err)
	db.containerDatabase.Close(t)
}

func (db *TestRedis) ConnectString(t *testing.T) string {
	return db.containerDatabase.ConnectionString(t)
}
