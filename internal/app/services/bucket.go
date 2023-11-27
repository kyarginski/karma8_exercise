package services

import (
	"log/slog"
	"net/http"
	"time"

	"karma8/internal/models"

	"github.com/google/uuid"
)

const requestPath = "/api/file"

type Bucket struct {
	log    *slog.Logger
	client *http.Client
	path   string
	ID     int64
}

func NewBucket(log *slog.Logger, path string, id int64) *Bucket {
	return &Bucket{
		log: log,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		path: path + requestPath,
		ID:   id,
	}
}

func (s *Bucket) GetFileItem(id uuid.UUID) (*models.FileItem, error) {
	return nil, nil
}

func (s *Bucket) PutFileItem(source *models.FileItem) (uuid.UUID, error) {
	return uuid.UUID{}, nil
}

func (s *Bucket) DeleteFileItem(id uuid.UUID) error {
	return nil
}

func (s *Bucket) GetBucketsInfo() ([]*models.ServerBucketInfo, error) {
	result := make([]*models.ServerBucketInfo, 0)
	result = append(result, &models.ServerBucketInfo{
		ID:      s.ID,
		Address: s.path,
	})

	return result, nil
}
