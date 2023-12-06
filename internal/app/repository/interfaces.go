package repository

import (
	"context"

	"karma8/internal/models"

	"github.com/google/uuid"
)

type IMetadata interface {
	GetFileMetadata(id uuid.UUID) (*models.MetadataItem, error)
	PutFileMetadata(source *models.MetadataItem) (uuid.UUID, error)
	DeleteFileMetadata(id uuid.UUID) error
}

type ICache interface {
	PutCacheItem(ctx context.Context, source *models.CacheItem) error
	GetCacheItem(ctx context.Context, checksum string) string
}

type IBucketInfo interface {
	GetBucketsInfo() ([]*models.ServerBucketInfo, error)
}

type IBucket interface {
	PutBucketItem(ctx context.Context, id string, source []byte) error
	GetBucketItem(ctx context.Context, id string) ([]byte, error)
}
