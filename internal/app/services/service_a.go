package services

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"karma8/internal/app/processes"
	"karma8/internal/app/repository"
	"karma8/internal/models"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

type ServiceA struct {
	log     *slog.Logger
	storage *repository.Storage
	buckets []*Bucket

	mu sync.Mutex
}

var (
	maxDateTime = time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC)
)

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
		log:     log,
		storage: storage,
		buckets: buckets,
	}, nil
}

// GetFileItem возвращает файл по его ID.
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

	data, err := s.GetFileFromBuckets(metadata.UUID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &models.FileItem{
		FileName:        metadata.FileName,
		FileContentType: metadata.ContentType,
		FileContent:     data,
	}, nil
}

// PutFileItem сохраняет файл на сервере и возвращает его ID.
func (s *ServiceA) PutFileItem(source *models.FileItem) (uuid.UUID, error) {
	const op = "serviceA.PutFileItem"

	path := processes.GetFileNameWithPathCache(source.FileName)
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
		ExpiredAt: time.Now().UTC().Add(3 * time.Minute), // TODO: в настройки.
	}
	err = s.storage.PutCacheItem(cache)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("%s: %w", op, err)
	}

	// Раскладываем файл по корзинам (buckets).
	err = s.PutFileIntoBuckets(newID, path)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("%s: %w", op, err)
	}

	return newID, nil
}

// DeleteFileItem удаляет файл по его ID.
func (s *ServiceA) DeleteFileItem(_ uuid.UUID) error {
	return nil
}

// GetBucketsInfo возвращает информацию о бакетах.
func (s *ServiceA) GetBucketsInfo() ([]*models.ServerBucketInfo, error) {
	return nil, nil
}

// GetFileNameFromCache возвращает имя файла из кэша по его ID.
func (s *ServiceA) GetFileNameFromCache(id uuid.UUID) string {
	item, err := s.storage.GetFileMetadata(id)
	if err != nil {
		return ""
	}

	return s.storage.GetCacheItem(item.Checksum)
}

// GetBucketsIDs возвращает ID всех бакетов.
func (s *ServiceA) GetBucketsIDs() []int64 {
	n := len(s.buckets)
	bucketIDs := make([]int64, n)

	for i, bucket := range s.buckets {
		bucketIDs[i] = bucket.ID
	}

	return bucketIDs
}

// Close закрывает соединение с БД.
func (s *ServiceA) Close() error {
	return s.storage.Close()
}

// PutFileIntoBuckets раскладывает файл по бакетам.
func (s *ServiceA) PutFileIntoBuckets(id uuid.UUID, path string) error {
	const op = "serviceA.PutFileIntoBuckets"

	items, err := processes.SplitFile(path, s.GetBucketsIDs())
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	var eg errgroup.Group

	// Обработка каждого элемента в отдельной горутине.
	for i, item := range items {
		item := item // создаем копию переменной item, чтобы избежать замыкания на изменяемой переменной в горутине
		i := i
		eg.Go(func() error {
			s.mu.Lock()
			defer s.mu.Unlock()

			s.buckets[i].log.Debug("SendToBucket",
				"id", id.String(),
				"bucketID", s.buckets[i].ID,
				"address", s.buckets[i].path,
			)
			return s.buckets[i].SendToBucket(&item, id)
		})
	}

	// Ожидание завершения всех горутин и проверка ошибок
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// GetFileFromBuckets собирает файл из бакетов.
func (s *ServiceA) GetFileFromBuckets(id uuid.UUID) ([]byte, error) {
	const op = "serviceA.GetFileFromBuckets"

	var eg errgroup.Group
	var mu sync.Mutex
	results := make(map[int64][]byte)

	// Запуск горутин для каждого бакета.
	for i, bucket := range s.buckets {
		bucket := bucket
		i := i
		eg.Go(func() error {
			s.mu.Lock()
			defer s.mu.Unlock()

			bucket.GetFromBucket(id, results, &mu)

			var res string
			if len(results[s.buckets[i].ID]) > 0 {
				res = string(results[s.buckets[i].ID][:10]) + "..."
			}

			s.buckets[i].log.Debug("GetFromBucket",
				"id", id.String(),
				"bucketID", s.buckets[i].ID,
				"address", s.buckets[i].path,
				"data", res,
			)

			return nil
		})
	}

	// Ожидание завершения всех горутин и проверка ошибок.
	if err := eg.Wait(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Объединение результатов в нужном порядке.
	var finalData []byte
	for _, bucket := range s.buckets {
		finalData = append(finalData, results[bucket.ID]...)
	}

	return finalData, nil
}

// ClearCache запускает периодическую очистку кэша.
func (s *ServiceA) ClearCache(d time.Duration) {
	// Создание таймера, который будет срабатывать каждые d интервалов.
	ticker := time.NewTicker(d)
	defer ticker.Stop()

	// Бесконечный цикл для периодического запуска функции очистки кэша.
	for range ticker.C {
		go s.runClearCache(time.Now().UTC())
	}
}

// runClearCache запускает очистку кэша.
func (s *ServiceA) runClearCache(current time.Time) {
	s.log.Debug("ClearCache " + current.String())

	// Получение списка файлов, которые нужно удалить.
	items, err := s.storage.GetExpiredCacheFilenames(current)
	if err != nil {
		s.log.Error("GetExpiredCacheFilenames", "error", err)
		return
	}
	// Удаление файлов и информации о них.
	err = s.storage.DeleteExpiredCacheFiles(items)
	if err != nil {
		s.log.Error("DeleteExpiredCacheFiles", "error", err)
		return
	}
	if len(items) > 0 {
		s.log.Debug("ClearCache deleted files", "count", len(items))
	}
}

func (s *ServiceA) ClearCacheAll() error {
	s.runClearCache(maxDateTime)

	return nil
}
