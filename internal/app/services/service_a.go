package services

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"karma8/internal/app/processes"
	"karma8/internal/app/repository"
	"karma8/internal/models"

	"github.com/google/uuid"
)

type ServiceA struct {
	storage *repository.Storage
	buckets []*Bucket
}

func NewServiceA(log *slog.Logger, connectString string) (*ServiceA, error) {
	const op = "serviceA.NewServiceA"

	storage, err := repository.New(connectString)
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

	return &ServiceA{
		storage: storage,
		buckets: buckets,
	}, nil
}

func (s *ServiceA) GetFileItem(id uuid.UUID) (*models.FileItem, error) {
	const op = "serviceA.GetFileItem"

	metadata, err := s.storage.GetFileMetadata(id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Проверяем наличие файла в кэше.
	fileName := s.storage.GetCacheItem(metadata.Checksum)
	if fileName != "" {
		// Если файл есть в кэше, то возвращаем его.
		data, err := processes.ReadFile(fileName)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		return &models.FileItem{
			FileName:        fileName,
			FileContentType: metadata.ContentType,
			FileContent:     data,
		}, nil
	}

	return nil, nil
}

func (s *ServiceA) PutFileItem(source *models.FileItem) (uuid.UUID, error) {
	const op = "serviceA.PutFileItem"

	path := filepath.Join(".", processes.PathCache, source.FileName)
	checksum, err := processes.CalculateChecksum(path)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("%s: %w", op, err)
	}

	metadata := &models.MetadataItem{
		UUID:        uuid.UUID{},
		Checksum:    checksum,
		FileName:    source.FileName,
		ContentType: source.FileContentType,
		BucketIDs:   s.GetBucketsIDs(),
	}
	// Сохраняем метаданные в БД.
	newID, err := s.storage.PutFileMetadata(metadata)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("%s: %w", op, err)
	}

	// Сохраняем файл в кэш.
	cache := &models.CacheItem{
		Checksum:  checksum,
		FileName:  source.FileName,
		ExpiredAt: time.Now().Add(3 * time.Minute), // TODO: в настройки.
	}
	err = s.storage.PutCacheItem(cache)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("%s: %w", op, err)
	}

	return newID, nil
}

func (s *ServiceA) DeleteFileItem(id uuid.UUID) error {
	return nil
}

func (s *ServiceA) GetBucketsInfo() ([]*models.ServerBucketInfo, error) {
	return nil, nil
}

func (s *ServiceA) PutFileItemToCache(source *models.CacheItem) error {
	return s.storage.PutCacheItem(source)
}

func (s *ServiceA) GetFileNameFromCache(id uuid.UUID) string {
	item, err := s.storage.GetFileMetadata(id)
	if err != nil {
		return ""
	}

	return s.storage.GetCacheItem(item.Checksum)
}

func (s *ServiceA) GetBucketsIDs() []int64 {
	n := len(s.buckets)
	bucketIDs := make([]int64, n)

	for i, bucket := range s.buckets {
		bucketIDs[i] = bucket.ID
	}

	return bucketIDs
}

func (s *ServiceA) Close() error {
	return s.storage.Close()
}
