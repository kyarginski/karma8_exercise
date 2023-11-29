package repository

import (
	"context"
	"fmt"

	"karma8/internal/models"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type StorageRedis struct {
	db *redis.Client
}

func (s *StorageRedis) Close() error {
	return s.db.Close()
}

func setDBRedis(path string, redisDB int) (*redis.Client, error) {
	opts, err := redis.ParseURL(path)
	if err != nil {
		return nil, err
	}
	opts.DB = redisDB
	rdb := redis.NewClient(opts)

	// Изменение конфигурации для сохранения данных на диск.
	ctx := context.Background()
	_, err = rdb.ConfigSet(ctx, "appendonly", "yes").Result()
	if err != nil {
		return nil, err
	}

	err = rdb.Ping(ctx).Err()
	if err != nil {
		return nil, err
	}

	return rdb, nil
}

func (s *StorageRedis) GetDB() *redis.Client {
	return s.db
}

func NewRedis(path string, redisDB int) (*StorageRedis, error) {
	const op = "repository.NewRedis"

	db, err := setDBRedis(path, redisDB)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &StorageRedis{db: db}, nil
}

func (s *StorageRedis) PutCacheItem(_ *models.CacheItem) error {
	return nil
}

func (s *StorageRedis) GetCacheItem(_ string) string {
	return ""
}

func (s *StorageRedis) GetFileMetadata(_ uuid.UUID) (*models.MetadataItem, error) {
	return nil, nil
}

func (s *StorageRedis) PutFileMetadata(_ *models.MetadataItem) (uuid.UUID, error) {
	return uuid.UUID{}, nil
}

func (s *StorageRedis) DeleteFileMetadata(_ uuid.UUID) error {
	return nil
}

func (s *StorageRedis) GetBucketsInfo() ([]*models.ServerBucketInfo, error) {
	return nil, nil
}

// PutBucketItem сохраняет часть файла в бакете.
func (s *StorageRedis) PutBucketItem(id string, source []byte) error {
	ctx := context.Background()
	err := s.db.Set(ctx, id, source, 0).Err()

	return err
}

// GetBucketItem возвращает часть файла из бакета по ID.
func (s *StorageRedis) GetBucketItem(id string) ([]byte, error) {
	ctx := context.Background()
	val, err := s.db.Get(ctx, id).Bytes()
	if err != nil {
		return nil, err
	}

	return val, nil
}
