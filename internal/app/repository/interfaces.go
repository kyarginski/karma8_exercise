package repository

import (
	"karma8/internal/models"

	"github.com/google/uuid"
)

type IMetadata interface {
	GetFileMetadata(id uuid.UUID) (*models.MetadataItem, error)
	PutFileMetadata(source *models.MetadataItem) (uuid.UUID, error)
	DeleteFileMetadata(id uuid.UUID) error
}

type ICache interface {
	PutCacheItem(source *models.CacheItem) error
	GetCacheItem(checksum string) string
}

type IBucket interface {
	GetBucketsInfo() ([]*models.ServerBucketInfo, error)
}
