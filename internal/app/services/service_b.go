package services

import (
	"fmt"
	"log/slog"

	"karma8/internal/app/repository"
	"karma8/internal/models"

	"github.com/google/uuid"
)

type ServiceB struct {
	log     *slog.Logger
	storage *repository.StorageRedis
	buckets []*Bucket
}

func NewServiceB(log *slog.Logger, connectString string, redisDB int) (*ServiceB, error) {
	const op = "serviceB.NewServiceB"

	storage, err := repository.NewRedis(connectString, redisDB)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	bucketsInfo, err := storage.GetBucketsInfo()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	n := len(bucketsInfo)
	buckets := make([]*Bucket, n)

	for i, bucketInfo := range bucketsInfo {
		buckets[i] = NewBucket(log, bucketInfo.Address, bucketInfo.ID)
	}

	return &ServiceB{
		log:     log,
		storage: storage,
		buckets: buckets,
	}, nil
}

func (s *ServiceB) GetFileItem(id uuid.UUID) (*models.FileItem, error) {
	const op = "serviceB.GetFileItem"

	// Получаем часть файла из БД.
	data, err := s.storage.GetBucketItem(id.String())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	item := &models.FileItem{
		ID:          id.String(),
		FileContent: data,
	}

	return item, nil
}

func (s *ServiceB) PutFileItem(source *models.FileItem) (uuid.UUID, error) {
	const op = "serviceB.PutFileItem"

	parsedUUID, err := uuid.Parse(source.ID)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("%s: %w", op, err)
	}

	// Сохраняем часть файла в БД.
	err = s.storage.PutBucketItem(source.ID, source.FileContent)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("%s: %w", op, err)
	}

	return parsedUUID, nil
}

func (s *ServiceB) DeleteFileItem(_ uuid.UUID) error {
	return nil
}

func (s *ServiceB) GetBucketsInfo() ([]*models.ServerBucketInfo, error) {
	return nil, nil
}

func (s *ServiceB) PutFileItemToCache(source *models.CacheItem) error {
	return s.storage.PutCacheItem(source)
}

func (s *ServiceB) GetFileNameFromCache(id uuid.UUID) string {
	item, err := s.storage.GetFileMetadata(id)
	if err != nil {
		return ""
	}

	return s.storage.GetCacheItem(item.Checksum)
}

func (s *ServiceB) GetBucketsIDs() []int64 {
	n := len(s.buckets)
	bucketIDs := make([]int64, n)

	for i, bucket := range s.buckets {
		bucketIDs[i] = bucket.ID
	}

	return bucketIDs
}

func (s *ServiceB) Close() error {
	return s.storage.Close()
}
