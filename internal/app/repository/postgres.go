package repository

import (
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

func (s *Storage) PutCacheItem(source *models.CacheItem) error {
	query := `
		INSERT INTO cache (checksum, filename, expired_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (checksum) DO UPDATE
		SET filename = EXCLUDED.filename, expired_at = EXCLUDED.expired_at;
	`
	_, err := s.db.Exec(query, source.Checksum, source.FileName, source.ExpiredAt)
	return err
}

func (s *Storage) GetCacheItem(checksum string) string {
	query := "SELECT filename FROM cache WHERE checksum = $1"

	var fileName string

	_ = s.db.QueryRow(query, checksum).Scan(&fileName)

	return fileName
}

func (s *Storage) GetFileMetadata(id uuid.UUID) (*models.MetadataItem, error) {
	query := "SELECT * FROM metadata WHERE uuid = $1"

	var item models.MetadataItem

	err := s.db.QueryRow(query, id).Scan(
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

func (s *Storage) PutFileMetadata(source *models.MetadataItem) (uuid.UUID, error) {
	// Генерация нового UUID
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
	err = s.db.QueryRow(
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

func (s *Storage) DeleteFileMetadata(id uuid.UUID) error {
	query := "DELETE FROM metadata WHERE uuid = $1"

	_, err := s.db.Exec(query, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) GetBucketsInfo() ([]*models.ServerBucketInfo, error) {
	query := `SELECT id, address FROM bucket WHERE active_sign = true order by id`

	rows, err := s.db.Query(query)
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

func (s *Storage) GetExpiredCacheFilenames() ([]models.CacheItem, error) {
	query := "SELECT filename, checksum FROM cache WHERE expired_at <= $1"

	current := time.Now().UTC()

	rows, err := s.db.Query(query, current)
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

func (s *Storage) DeleteExpiredCacheFiles(items []models.CacheItem) error {
	query := "DELETE FROM cache WHERE checksum = $1"

	for _, item := range items {
		_, err := s.db.Exec(query, item.Checksum)
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
