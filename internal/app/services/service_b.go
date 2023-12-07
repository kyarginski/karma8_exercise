package services

import (
	"context"
	"fmt"
	"log/slog"

	"karma8/internal/app/repository"
	trccontext "karma8/internal/lib/context"
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

// GetFileItem возвращает файл по его ID.
func (s *ServiceB) GetFileItem(ctx context.Context, id uuid.UUID) (*models.FileItem, error) {
	const op = "serviceB.GetFileItem"

	ctx, span := trccontext.WithTelemetrySpan(ctx, op)
	defer span.End()

	// Получаем часть файла из БД.
	data, err := s.storage.GetBucketItem(ctx, id.String())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	item := &models.FileItem{
		ID:          id.String(),
		FileContent: data,
	}

	return item, nil
}

// PutFileItem сохраняет файл в БД и возвращает его ID.
func (s *ServiceB) PutFileItem(ctx context.Context, source *models.FileItem) (uuid.UUID, error) {
	const op = "serviceB.PutFileItem"

	ctx, span := trccontext.WithTelemetrySpan(ctx, op)
	defer span.End()

	parsedUUID, err := uuid.Parse(source.ID)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("%s: %w", op, err)
	}

	// Сохраняем часть файла в БД.
	err = s.storage.PutBucketItem(ctx, source.ID, source.FileContent)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("%s: %w", op, err)
	}

	return parsedUUID, nil
}

// DeleteFileItem удаляет файл по его ID.
func (s *ServiceB) DeleteFileItem(_ context.Context, _ uuid.UUID) error {
	return nil
}

// GetBucketsInfo возвращает информацию об активных бакетах.
func (s *ServiceB) GetBucketsInfo() ([]*models.ServerBucketInfo, error) {
	return nil, nil
}

// PutFileItemToCache сохраняет файл в кэше.
func (s *ServiceB) PutFileItemToCache(source *models.CacheItem) error {
	return s.storage.PutCacheItem(source)
}

// GetFileNameFromCache возвращает имя файла из кэша.
func (s *ServiceB) GetFileNameFromCache(id uuid.UUID) string {
	item, err := s.storage.GetFileMetadata(id)
	if err != nil {
		return ""
	}

	return s.storage.GetCacheItem(item.Checksum)
}

// GetBucketsIDs возвращает ID бакетов.
func (s *ServiceB) GetBucketsIDs() []int64 {
	n := len(s.buckets)
	bucketIDs := make([]int64, n)

	for i, bucket := range s.buckets {
		bucketIDs[i] = bucket.ID
	}

	return bucketIDs
}

// Close закрывает соединение с БД.
func (s *ServiceB) Close() error {
	return s.storage.Close()
}

func (s *ServiceB) ClearCacheAll() error {
	return nil
}
