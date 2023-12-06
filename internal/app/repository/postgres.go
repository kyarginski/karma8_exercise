package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"karma8/internal/app/processes"
	"karma8/internal/models"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func setDB(path string) (*sql.DB, error) {
	db, err := sql.Open("postgres", path)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (s *Storage) GetDB() *sql.DB {
	return s.db
}

func New(path string) (*Storage, error) {
	const op = "repository.New"

	db, err := setDB(path)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

// PutCacheItem сохраняет информацию о файле в кэше в БД.
func (s *Storage) PutCacheItem(ctx context.Context, source *models.CacheItem) error {
	query := `
		INSERT INTO cache (checksum, filename, expired_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (checksum) DO UPDATE
		SET filename = EXCLUDED.filename, expired_at = EXCLUDED.expired_at;
	`
	_, err := s.db.ExecContext(ctx, query, source.Checksum, source.FileName, source.ExpiredAt)
	return err
}

// GetCacheItem возвращает имя файла из кэша по его контрольной сумме.
func (s *Storage) GetCacheItem(ctx context.Context, checksum string) string {
	query := "SELECT filename FROM cache WHERE checksum = $1"

	var fileName string

	_ = s.db.QueryRowContext(ctx, query, checksum).Scan(&fileName)

	return fileName
}

// GetFileMetadata возвращает метаданные файла по UUID.
func (s *Storage) GetFileMetadata(ctx context.Context, id uuid.UUID) (*models.MetadataItem, error) {
	query := "SELECT * FROM metadata WHERE uuid = $1"

	var item models.MetadataItem

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&item.UUID,
		&item.Checksum,
		&item.FileName,
		&item.ContentType,
		pq.Array(&item.BucketIDs),
		&item.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

// PutFileMetadata сохраняет метаданные файла в БД - возвращает новый UUID файла.
func (s *Storage) PutFileMetadata(ctx context.Context, source *models.MetadataItem) (uuid.UUID, error) {
	// Генерация нового UUID.
	newUUID, err := uuid.NewUUID()
	if err != nil {
		return uuid.Nil, err
	}

	// Подготовка запроса INSERT
	query := `
		INSERT INTO metadata (uuid, checksum, filename, content_type, bucket_ids)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (checksum) DO UPDATE
		SET uuid = $1, filename = $3, content_type = $4, bucket_ids = $5
		RETURNING uuid;
	`

	// Выполнение запроса
	err = s.db.QueryRowContext(
		ctx,
		query,
		newUUID,
		source.Checksum,
		source.FileName,
		source.ContentType,
		pq.Array(source.BucketIDs),
	).Scan(&newUUID)

	if err != nil {
		return uuid.Nil, err
	}

	return newUUID, nil
}

// DeleteFileMetadata удаляет метаданные файла по UUID.
func (s *Storage) DeleteFileMetadata(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM metadata WHERE uuid = $1"

	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

// GetBucketsInfo возвращает информацию о всех активных бакетах.
func (s *Storage) GetBucketsInfo(ctx context.Context) ([]*models.ServerBucketInfo, error) {
	query := `SELECT id, address FROM bucket WHERE active_sign = true order by id`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	buckets := make([]*models.ServerBucketInfo, 0)

	for rows.Next() {
		var bucket models.ServerBucketInfo

		err = rows.Scan(&bucket.ID, &bucket.Address)
		if err != nil {
			return nil, err
		}

		buckets = append(buckets, &bucket)
	}

	return buckets, nil
}

// GetExpiredCacheFilenames возвращает информацию о файлах из кэша, которые просрочены.
func (s *Storage) GetExpiredCacheFilenames(ctx context.Context, current time.Time) ([]models.CacheItem, error) {
	query := "SELECT filename, checksum FROM cache WHERE expired_at <= $1"

	rows, err := s.db.QueryContext(ctx, query, current)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cacheItems := make([]models.CacheItem, 0)

	for rows.Next() {
		var fileName, checksum string

		if err := rows.Scan(&fileName, &checksum); err != nil {
			return nil, err
		}

		cacheItems = append(cacheItems, models.CacheItem{
			Checksum: checksum,
			FileName: fileName,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cacheItems, nil
}

// DeleteExpiredCacheFiles удаляет файлы из кэша, которые просрочены.
func (s *Storage) DeleteExpiredCacheFiles(ctx context.Context, items []models.CacheItem) error {
	query := "DELETE FROM cache WHERE checksum = $1"

	for _, item := range items {
		_, err := s.db.ExecContext(ctx, query, item.Checksum)
		if err != nil {
			return err
		}

		// Удаление файла из кэша.
		err = processes.DeleteFile(item.FileName)
		if err != nil {
			return err
		}
	}

	return nil
}
